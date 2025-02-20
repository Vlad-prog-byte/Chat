package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*Client]bool) // Храним клиентов
	clientsMu sync.Mutex               // Мьютекс для безопасного изменения карты клиентов
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// Структура клиента с каналом для сообщений
type Client struct {
	conn *websocket.Conn
	send chan []byte // Канал для отправки сообщений клиенту
}

// Структура сообщения
type Message struct {
	Text string `json:"text"` // Должно быть экспортируемым (с заглавной буквы), иначе JSON не распарсит
}

// Горутина для отправки сообщений клиенту (читаем из `send` и пишем в WebSocket)
func (c *Client) writePump() {
	defer func() {
		c.conn.Close() // Закрываем соединение при выходе
		clientsMu.Lock()
		delete(clients, c) // Удаляем клиента из списка
		clientsMu.Unlock()
	}()

	for msg := range c.send { // Читаем из канала send
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("Ошибка отправки:", err)
			return // Если ошибка, выходим из горутины
		}
	}
}

// Функция для отправки сообщения всем клиентам
func sendMessage(msg Message) {
	b, err := json.Marshal(msg)
	if err != nil {
		log.Println("Ошибка сериализации:", err)
		return
	}

	clientsMu.Lock()
	for client := range clients {
		client.send <- b // Вместо прямого WriteMessage отправляем в канал клиента
	}
	clientsMu.Unlock()
}

// Обработчик WebSocket-соединения
func createConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка апгрейда:", err)
		http.Error(w, "Не удалось установить соединение", http.StatusInternalServerError)
		return
	}

	client := &Client{
		conn: ws,
		send: make(chan []byte, 10), // Буферизированный канал для отправки сообщений
	}

	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()

	// Запускаем writePump в отдельной горутине
	go client.writePump()

	// Читаем сообщения от клиента
	for {
		mt, b, err := ws.ReadMessage()
		if err != nil || mt == websocket.CloseMessage {
			log.Println("Ошибка чтения или клиент отключился:", err)
			return
		}

		var msg Message
		if err = json.Unmarshal(b, &msg); err != nil {
			log.Println("Ошибка парсинга JSON:", err)
			ws.WriteMessage(websocket.TextMessage, []byte(`{"error": "invalid message format"}`))
			return
		}

		sendMessage(msg) // Отправляем сообщение всем клиентам
	}
}

func main() {
	http.HandleFunc("/ws", createConnection)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Panic(err)
	}
}
