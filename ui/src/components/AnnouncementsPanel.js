import React, { useState } from 'react';

const AnnouncementsPanel = ({ currentUser, messages, ws }) => {
  const [input, setInput] = useState('');

  const sendAnnouncement = () => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      const msg = {
        type: 'announcement',
        sender: currentUser,
        timestamp: new Date(),
        content: input
      };
      ws.send(JSON.stringify(msg));
      setInput('');
    }
  };

  return (
    <div>
      <h2>Announcements</h2>
      <div style={{ border: '1px solid #ccc', padding: '10px', height: '250px', overflowY: 'scroll' }}>
        {messages.map((msg, idx) => (
          <div key={idx}>
            <strong>{msg.sender}:</strong> {msg.content}
          </div>
        ))}
      </div>
      <div style={{ marginTop: '10px' }}>
        <input
          type="text"
          placeholder="Enter announcement"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          style={{ width: '80%', marginRight: '10px' }}
        />
        <button onClick={sendAnnouncement}>Send</button>
      </div>
    </div>
  );
};

export default AnnouncementsPanel;
