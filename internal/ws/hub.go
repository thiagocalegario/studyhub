package ws

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	UserID    int
	UserName  string
	OnMessage func(c *Client, data []byte)
}

type Hub struct {
	disciplineID int
	clients      map[*Client]bool
	broadcast    chan []byte
	register     chan *Client
	unregister   chan *Client
}

var (
	hubs   = make(map[int]*Hub)
	hubsMu sync.RWMutex
)

func GetHub(disciplineID int) *Hub {
	hubsMu.RLock()
	h, ok := hubs[disciplineID]
	hubsMu.RUnlock()
	if ok {
		return h
	}

	hubsMu.Lock()
	defer hubsMu.Unlock()
	if h, ok := hubs[disciplineID]; ok {
		return h
	}

	h = &Hub{
		disciplineID: disciplineID,
		clients:      make(map[*Client]bool),
		broadcast:    make(chan []byte, 256),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
	}
	hubs[disciplineID] = h
	go h.Run()
	return h
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case data := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) BroadcastJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	h.broadcast <- data
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return upgrader.Upgrade(w, r, nil)
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		if c.OnMessage != nil {
			c.OnMessage(c, message)
		}
	}
}

func (c *Client) WritePump() {
	defer c.conn.Close()

	for data := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			break
		}
	}
}
