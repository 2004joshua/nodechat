import React, { useEffect, useState, useRef } from 'react';
import './App.css';

function App() {
  const queryParams = new URLSearchParams(window.location.search);
  const initialUsername = queryParams.get('username') || "WebUser";

  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const [username] = useState(initialUsername);
  const [topicInput, setTopicInput] = useState("");
  const [subscribedTopics, setSubscribedTopics] = useState([]);
  const fileRef = useRef(null);

  const fetchMessages = async () => {
    try {
      const res = await fetch("/messages");
      const data = await res.json();
      if (!Array.isArray(data)) return;

      if (Notification.permission === "granted") {
        data.forEach((msg) => {
          if (msg.type === "notification") {
            new Notification(`${msg.sender}`, { body: msg.content });
          }
        });
      }

      setMessages(data.filter(msg => msg.type !== "notification"));
    } catch (err) {
      console.error("Error fetching messages:", err);
    }
  };

  const fetchSubscriptions = async () => {
    try {
      const res = await fetch("/subscriptions");
      const data = await res.json();
      setSubscribedTopics(data || []);
    } catch (err) {
      console.error("Error fetching subscriptions:", err);
    }
  };

  const handleSubscribe = async () => {
    if (topicInput.trim()) {
      await fetch(`/subscribe?topic=${topicInput}`);
      setSubscribedTopics(prev => [...new Set([...prev, topicInput])]);
      setTopicInput("");
    }
  };

  const handleUnsubscribe = async () => {
    if (topicInput.trim()) {
      await fetch(`/unsubscribe?topic=${topicInput}`);
      setSubscribedTopics(prev => prev.filter(t => t !== topicInput));
      setTopicInput("");
    }
  };

  useEffect(() => {
    if (Notification.permission !== "granted") {
      Notification.requestPermission();
    }
    fetchMessages();
    fetchSubscriptions();
    const interval = setInterval(fetchMessages, 3000);
    return () => clearInterval(interval);
  }, []);

  const handleSend = async () => {
    // 1) GIF upload branch
    if (fileRef.current?.files.length) {
      const file = fileRef.current.files[0];
      const form = new FormData();
      form.append("file", file);

      try {
        const uploadRes = await fetch("/upload", {
          method: "POST",
          body: form,
        });
        const { url, fileName } = await uploadRes.json();

        const payload = {
          type: "file",
          sender: username,
          fileUrl: url,
          fileName,
          topic: ""
        };

        await fetch("/messages", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload)
        });

        fileRef.current.value = "";
        fetchMessages();
      } catch (err) {
        console.error("ERROR: Uploading GIF didn't work", err);
      }
      return;
    }

    // 2) Text/chat branch
    if (newMessage.trim() === "") return;

    let content = newMessage;
    let topic = "";

    if (newMessage.startsWith("/topic ")) {
      const parts = newMessage.split(" ");
      if (parts.length >= 3) {
        topic = parts[1];
        content = parts.slice(2).join(" ");
      }
    }

    const messagePayload = {
      type: "chat",
      sender: username,
      content,
      topic
    };

    try {
      const res = await fetch("/messages", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(messagePayload)
      });
      if (res.ok) {
        setNewMessage("");
        fetchMessages();
      } else {
        console.error("Failed to send message");
      }
    } catch (err) {
      console.error("Error sending message", err);
    }
  };

  return (
    <div className="app-container">
      <div className="sidebar">
        <h2>{username}</h2>

        <input
          type="text"
          placeholder="Enter topic"
          value={topicInput}
          onChange={e => setTopicInput(e.target.value)}
        />
        <button onClick={handleSubscribe}>Subscribe</button>
        <button onClick={handleUnsubscribe}>Unsubscribe</button>

        <h4>Subscribed Topics:</h4>
        {subscribedTopics.length > 0 ? (
          <ul>
            {subscribedTopics.map((topic, i) => <li key={i}>{topic}</li>)}
          </ul>
        ) : (
          <p className="none">None</p>
        )}
      </div>

      <div className="chat-container">
        <div className="messages">
          {messages.map((msg, i) => (
            <div
              key={i}
              className={`message-bubble ${msg.sender === username ? 'self' : 'other'}`}
            >
              <div className="message-header">
                <strong>{msg.sender}</strong>
                <span className="timestamp">
                  {new Date(msg.timestamp * 1000).toLocaleTimeString()}
                </span>
              </div>
              <div className="message-content">
                {msg.type === 'file' && msg.fileUrl ? (
                  <img
                    src={msg.fileUrl}
                    alt={msg.fileName}
                    className="max-w-xs rounded-lg shadow"
                  />
                ) : (
                  msg.content
                )}
              </div>
            </div>
          ))}
        </div>

        <div className="input-area">
          <input
            type="text"
            placeholder="Type your message..."
            value={newMessage}
            onChange={e => setNewMessage(e.target.value)}
          />
          <input
            type="file"
            accept="image/gif"
            ref={fileRef}
            className="ml-2"
          />
          <button onClick={handleSend}>Send</button>
        </div>
      </div>
    </div>
  );
}

export default App;
