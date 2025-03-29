import React, { useState } from 'react';

const ChatPanel = ({ currentUser, recipient, messages, ws, onBack }) => {
  const [input, setInput] = useState('');

  const sendMessage = () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      const msg = {
        type: 'chat',
        sender: currentUser,
        recipient: recipient.username,
        timestamp: new Date(),
        content: input
      };
      ws.send(JSON.stringify(msg));
      setInput('');
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div>
        <button onClick={onBack}>Back</button>
        <h2>Chat with {recipient.username}</h2>
      </div>
      <div style={{ flex: 1, border: '1px solid #ccc', overflowY: 'scroll', marginBottom: '10px', padding: '10px' }}>
        {messages.map((msg, idx) => (
          <div key={idx}>
            <strong>{msg.sender}:</strong> {msg.content}
          </div>
        ))}
      </div>
      <div>
        <input
          type="text"
          placeholder={`Message ${recipient.username}`}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          style={{ width: '80%', marginRight: '10px' }}
        />
        <button onClick={sendMessage}>Send</button>
      </div>
    </div>
  );
};

export default ChatPanel;
