import React, { useState, useEffect } from 'react';
import AnnouncementsPanel from './AnnouncementsPanel';
import UserList from './UserList';
import ChatPanel from './ChatPanel';

const ChatApp = ({ currentUser }) => {
  // Hard-coded user list for demonstration
  const [userList] = useState([
    { username: 'peer2' },
    { username: 'peer3' }
  ]);

  const [selectedUser, setSelectedUser] = useState(null);
  const [messages, setMessages] = useState([]);
  const [ws, setWs] = useState(null);

  useEffect(() => {
    // Connect to a local WebSocket server on port 8080 (adjust as needed)
    const socket = new WebSocket('ws://localhost:8080/ws');

    socket.onopen = () => console.log('WebSocket connected');
    socket.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        setMessages(prev => [...prev, msg]);
      } catch (e) {
        console.error('Invalid JSON from WS:', event.data);
      }
    };
    socket.onclose = () => console.log('WebSocket disconnected');

    setWs(socket);
    return () => socket.close();
  }, []);

  return (
    <div style={{ display: 'flex', height: '100vh' }}>
      <div style={{ width: '200px', borderRight: '1px solid #ccc' }}>
        <UserList currentUser={currentUser} users={userList} onSelect={setSelectedUser} />
      </div>
      <div style={{ flex: 1, padding: '20px' }}>
        {!selectedUser ? (
          <AnnouncementsPanel
            currentUser={currentUser}
            messages={messages.filter(m => m.type === 'announcement')}
            ws={ws}
          />
        ) : (
          <ChatPanel
            currentUser={currentUser}
            recipient={selectedUser}
            messages={messages.filter(m =>
              (m.sender === selectedUser.username && m.recipient === '') ||
              (m.recipient === selectedUser.username && m.type === 'chat') ||
              (m.sender === currentUser && m.recipient === selectedUser.username)
            )}
            ws={ws}
            onBack={() => setSelectedUser(null)}
          />
        )}
      </div>
    </div>
  );
};

export default ChatApp;
