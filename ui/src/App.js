// ui/src/App.js
import React, { useState, useEffect, useRef } from 'react';
import './App.css';

function App() {
  const [messages, setMessages] = useState([]);
  const [text, setText] = useState('');
  const ws = useRef(null);

  useEffect(() => {
    ws.current = new WebSocket(`ws://${window.location.host}/ws`);
    ws.current.onmessage = ({ data }) => {
      const { sender, text } = JSON.parse(data);
      setMessages(prev => [...prev, { sender, text }]);
    };
    return () => ws.current.close();
  }, []);

  const send = () => {
    if (!text.trim()) return;
    ws.current.send(JSON.stringify({ sender: 'You', text }));
    setMessages(prev => [...prev, { sender: 'You', text }]);
    setText('');
  };

  return (
    <div className="chat-container">
      <div className="messages">
        {messages.map((m, i) => (
          <div key={i}><strong>{m.sender}:</strong> {m.text}</div>
        ))}
      </div>
      <div className="input-bar">
        <input
          value={text}
          onChange={e => setText(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && send()}
          placeholder="Type a message..."
        />
        <button onClick={send}>Send</button>
      </div>
    </div>
  );
}

export default App;