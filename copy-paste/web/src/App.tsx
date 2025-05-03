import { useState, useEffect } from 'react';
import './App.css';

function App() {
  const [text, setText] = useState('');
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState('');
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [copied, setCopied] = useState(false);

  // Initialize WebSocket connection
  useEffect(() => {
    // Use secure WebSocket in production
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.host}/ws`;
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('Connected to the server');
      setStatus('Connected');
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'text_update') {
        setText(data.text);
      } else if (data.type === 'code_assigned') {
        setCode(data.code);
        setLoading(false);
      }
    };

    ws.onclose = () => {
      console.log('Disconnected from the server');
      setStatus('Disconnected');
    };

    setSocket(ws);

    return () => {
      ws.close();
    };
  }, []);

  const generateNewCode = () => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      setLoading(true);
      socket.send(JSON.stringify({ type: 'generate_code' }));
    }
  };

  const joinWithCode = (inputCode: string) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type: 'join_code', code: inputCode }));
      setCode(inputCode);
    }
  };

  const handleTextChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newText = e.target.value;
    setText(newText);

    // Send text update to server
    if (socket && socket.readyState === WebSocket.OPEN && code) {
      socket.send(JSON.stringify({
        type: 'update_text',
        code: code,
        text: newText
      }));
    }
  };

  const handleCodeInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const inputCode = e.target.value.toUpperCase();
    setCode(inputCode);
  };

  const handleCodeSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    joinWithCode(code);
  };

  const copyToClipboard = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="flex flex-col min-h-screen bg-slate-900">
      <header className="py-6 bg-teal-900 text-white">
        <div className="container mx-auto px-4">
          <h1 className="text-2xl font-bold">CopyPaste</h1>
          <p className="text-teal-200">Share text between your devices instantly</p>
        </div>
      </header>

      <main className="flex-grow container mx-auto px-4 py-8">
        <div className="bg-slate-800 rounded-lg shadow-xl p-6 mb-6">
          <div className="mb-6">
            <h2 className="text-lg font-semibold mb-2 text-white">Connection Code</h2>

            {!code ? (
              <div className="flex flex-col md:flex-row gap-4">
                <button
                  onClick={generateNewCode}
                  className="px-4 py-2 bg-teal-600 text-white rounded hover:bg-teal-700 transition-colors"
                  disabled={loading}
                >
                  {loading ? 'Generating...' : 'Generate New Code'}
                </button>

                <form onSubmit={handleCodeSubmit} className="flex flex-1 gap-2">
                  <input
                    type="text"
                    placeholder="Or enter existing code"
                    value={code}
                    onChange={handleCodeInputChange}
                    className="flex-1 px-3 py-2 bg-slate-700 text-white border border-slate-600 rounded focus:outline-none focus:ring-2 focus:ring-teal-500"
                    maxLength={6}
                  />
                  <button
                    type="submit"
                    className="px-4 py-2 bg-slate-700 text-white rounded hover:bg-slate-600 transition-colors"
                  >
                    Join
                  </button>
                </form>
              </div>
            ) : (
              <div className="flex items-center gap-4">
                <span className="text-xl font-mono bg-slate-700 text-white px-4 py-2 rounded">{code}</span>
                <span className="text-sm text-slate-300">Use this code on your other device</span>
                <button
                  onClick={() => setCode('')}
                  className="text-sm text-slate-400 hover:text-slate-200"
                >
                  Change
                </button>
              </div>
            )}
          </div>

          <div>
            <div className="flex justify-between items-center mb-2">
              <h2 className="text-lg font-semibold text-white">Shared Text</h2>
              <button
                onClick={copyToClipboard}
                className="px-3 py-1 text-sm bg-slate-700 text-white hover:bg-slate-600 rounded transition-colors"
              >
                {copied ? 'Copied!' : 'Copy to Clipboard'}
              </button>
            </div>
            <textarea
              value={text}
              onChange={handleTextChange}
              placeholder="Type or paste your text here..."
              className="w-full h-64 p-3 bg-slate-700 text-white border border-slate-600 rounded resize-none focus:outline-none focus:ring-2 focus:ring-teal-500 placeholder-slate-400"
            />
          </div>
        </div>

        <div className="text-center text-sm text-slate-400">
          <p>Status: {status}</p>
        </div>
      </main>

      <footer className="py-4 bg-slate-800">
        <div className="container mx-auto px-4 text-center text-slate-400 text-sm">
          CopyPaste - Simple text sharing between devices
        </div>
      </footer>
    </div>
  );
}

export default App;