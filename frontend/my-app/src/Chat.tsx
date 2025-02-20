import React, { useState } from "react";
import useWebSocket, { ReadyState } from "react-use-websocket";

const WS_URL = "ws://localhost:8080/ws"

const Chat: React.FC = () => {
  const [message, setMessage] = useState("")
  const [chatHistory, setChatHistory] = useState<string[]>([])


  const { sendJsonMessage, lastJsonMessage, readyState } = useWebSocket(WS_URL, {
    onOpen: () => console.log("Connected to WebSocket"),
    onMessage: (event) => {
      const msg = JSON.parse(event.data);
      setChatHistory((prev) => [...prev, msg.text])
    },
    onClose: () => console.log("WebSocket Disconnected"),
    shouldReconnect: () => true
  })

  const sendMessage = () => {
    if (message.trim() !== "") {
      sendJsonMessage({ text: message })
      setMessage("")
    }
  }

  const handleKeypress = (e: React.KeyboardEvent<HTMLElement>) => {
    if (e.keyCode === 13) {
      sendMessage()
    }
  };

  return (
    <div className="container mt-4">
      <h2>Чат</h2>
      <div className="chat-box border p-3 mb-3" style={{ height: "300px", overflowY: "auto" }}>
        {chatHistory.map((msg, index) => (
          <div key={index} className="alert alert-primary">
            {msg}
          </div>
        ))}
      </div>
      <div className="input-group">
        <input
          type="text"
          className="form-control"
          placeholder="Введите сообщение..."
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          onKeyDown={handleKeypress}
        />
        <button
          className="btn btn-primary"
          onClick={sendMessage}
          disabled={readyState !== ReadyState.OPEN}
        >
          Отправить
        </button>
      </div>
      {readyState !== ReadyState.OPEN && <p className="text-danger mt-2">Соединение с сервером потеряно...</p>}
    </div>
  );
};

export default Chat;
