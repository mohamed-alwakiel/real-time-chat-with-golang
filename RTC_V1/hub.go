package main

type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Incoming messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		// عميل جديد دخل
		case client := <-h.register:
			h.clients[client] = true
		// عميل خرج أو فصل
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

			// رسالة وصلت من عميل وعايزين نبعِتها للكل
		// رسالة وصلت من عميل وعايزين نبعِتها للكل
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// لو فيه مشكلة عند العميل، شيله
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
