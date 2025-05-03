# copy-paste
copy anywhere, paste anywhere. A minimal web application for quickly sharing text between devices. No need to email yourself or use messaging apps(At least that's what I usually do lol...)

## Features

- **Simple Text Sharing**: Copy text on one device, access it on another
- **Real-time Updates**: Changes sync instantly across all connected devices
- **Unique Codes**: Generate short codes to connect your devices
- **Minimal UI**: Clean, responsive interface focused on the task
- **Privacy**: No account required, no text storage on servers after disconnection

## Tech Stack

- **Frontend**: React with TypeScript and Vite
- **Styling**: TailwindCSS
- **Backend**: Golang with WebSockets
- **Deployment**: Containerized with Docker

## Project Structure

```
copy-paste/
├── web/            # React + TypeScript frontend
│   ├── src/
│   │   ├── App.tsx      # Main application component
│   │   ├── App.css      # Styles with TailwindCSS
│   │   └── main.tsx     # Entry point
│   ├── public/          # Static assets
│   ├── package.json     # Frontend dependencies
│   └── vite.config.ts   # Vite configuration
├── main.go              # Golang backend server
├── go.mod               # Go dependencies
├── go.sum               # Go dependency checksums
├── Dockerfile           # Container configuration
├── build.sh            # Build script
└── run.sh              # Run script
```


## How to Use

1. **Setup Your Environment**:
   - Make sure you have Go installed.
   - Make sure you have Node.js installed.

2. **Installation**:
   ```bash
   # Clone the repository
   git clone https://github.com/vantage-ola/copy-paste.git
   cd copy-paste

   # Run the setup script
   ./setup.sh
   ```

3. **Development**:
   ```bash
   # Start the frontend development server
   cd web
   npm run dev

   # In another terminal, start the backend server
   go run main.go
   ```

4. **Production Build**:
   ```bash
   # Build the entire application
   ./build.sh

   # Run the server
   ./copy-paste
   ```

5. **Using Docker**:
   ```bash
   # Build the Docker image
   docker build -t copy-paste .

   # Run the container
   docker run -p 8080:8080 copy-paste
   ```

## Using the App

1. Open the app on your first device (e.g., your computer) at `http://localhost:8080`
2. Click "Generate New Code" to create a unique 6-character code
3. Type or paste the text you want to share
4. On your second device (e.g., your phone), open the app and enter the same code
5. The text will be synchronized between devices in real-time
6. Click "Copy to Clipboard" to copy the text on your second device

## Future Improvements

- End-to-end encryption for enhanced privacy
- Optional persistent sessions for longer-term sharing
- File sharing capabilities
- Mobile app versions
