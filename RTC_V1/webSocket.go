package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// السماح بكل origins (علشان نسهّل الدنيا مؤقتًا)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWs: النقطة اللي بتحوّل HTTP → WebSocket
func ServerWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256), // قناة إرسال للعميل
	}

	client.hub.register <- client // سجّل العميل في الـ Hub

	// شغّل Goroutines للقراءة والكتابة
	go client.writePump()
	go client.readPump()
}
