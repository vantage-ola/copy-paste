package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string `json:"type"`
	Code string `json:"code,omitempty"`
	Text string `json:"text,omitempty"`
}

// connnected client
type Client struct {
	conn *websocket.Conn
	send chan []byte
	code string
}

type TextSession struct {
	text    string
	clients map[*Client]bool
}

var (
	// Upgrader for WebSocket connections
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all connections in development
		},
	}

	// Global state
	sessions     = make(map[string]*TextSession)
	clients      = make(map[*Client]bool)
	mutex        sync.RWMutex
	registerCh   = make(chan *Client)
	unregisterCh = make(chan *Client)
	broadcastCh  = make(chan Message)
)

// GenerateUniqueCode creates a new unique 6-character code
func GenerateUniqueCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	const codeLength = 6

	for {
		code := make([]byte, codeLength)
		for i := range code {
			code[i] = charset[rand.Intn(len(charset))]
		}
		codeStr := string(code)

		// Check if code already exists
		mutex.RLock()
		_, exists := sessions[codeStr]
		mutex.RUnlock()

		if !exists {
			return codeStr
		}
	}
}

// ServeWs handles WebSocket connections
func ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	registerCh <- client

	// Start goroutines for reading and writing
	go client.readPump()
	go client.writePump()
}

// readPump handles incoming messages from the client
func (c *Client) readPump() {
	defer func() {
		unregisterCh <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB limit
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error decoding message: %v", err)
			continue
		}

		switch msg.Type {
		case "generate_code":
			newCode := GenerateUniqueCode()
			c.code = newCode

			mutex.Lock()
			sessions[newCode] = &TextSession{
				text:    "",
				clients: make(map[*Client]bool),
			}
			sessions[newCode].clients[c] = true
			mutex.Unlock()

			response := Message{
				Type: "code_assigned",
				Code: newCode,
			}
			responseJSON, _ := json.Marshal(response)
			c.send <- responseJSON

		case "join_code":
			codeToJoin := msg.Code

			mutex.Lock()
			session, exists := sessions[codeToJoin]
			if exists {
				// Remove from previous session if any
				if c.code != "" && c.code != codeToJoin {
					if prevSession, ok := sessions[c.code]; ok {
						delete(prevSession.clients, c)
						// Clean up empty sessions
						if len(prevSession.clients) == 0 {
							delete(sessions, c.code)
						}
					}
				}

				c.code = codeToJoin
				session.clients[c] = true

				// Send current text to the client
				textUpdate := Message{
					Type: "text_update",
					Text: session.text,
				}
				textUpdateJSON, _ := json.Marshal(textUpdate)
				c.send <- textUpdateJSON
			} else {
				// Create new session if it doesn't exist
				sessions[codeToJoin] = &TextSession{
					text:    "",
					clients: make(map[*Client]bool),
				}
				sessions[codeToJoin].clients[c] = true
				c.code = codeToJoin
			}
			mutex.Unlock()

		case "update_text":
			if msg.Code == "" {
				continue
			}

			mutex.Lock()
			session, exists := sessions[msg.Code]
			if exists {
				session.text = msg.Text

				// Broadcast to all clients in the session
				for client := range session.clients {
					if client != c { // Don't send back to the sender
						update := Message{
							Type: "text_update",
							Text: msg.Text,
						}
						updateJSON, _ := json.Marshal(update)
						client.send <- updateJSON
					}
				}
			}
			mutex.Unlock()
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Hub manages clients and broadcasts
func Hub() {
	for {
		select {
		case client := <-registerCh:
			mutex.Lock()
			clients[client] = true
			mutex.Unlock()

		case client := <-unregisterCh:
			mutex.Lock()
			if _, ok := clients[client]; ok {
				delete(clients, client)
				close(client.send)

				// Remove from session
				if client.code != "" {
					if session, ok := sessions[client.code]; ok {
						delete(session.clients, client)

						// Clean up empty sessions
						if len(session.clients) == 0 {
							delete(sessions, client.code)
						}
					}
				}
			}
			mutex.Unlock()
		}
	}
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Start the hub in a goroutine
	go Hub()

	// File server for frontend assets
	fs := http.FileServer(http.Dir("./dist"))
	http.Handle("/", fs)

	// WebSocket endpoint
	http.HandleFunc("/ws", ServeWs)

	// Start server
	port := "8080"
	fmt.Printf("Server starting on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
