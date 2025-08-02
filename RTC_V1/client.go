package main

import (
	"bytes"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	// الوقت المسموح به لإغلاق الاتصال
	writeWait = 10 * time.Second

	// الوقت المسموح فيه بانتظار رسالة من العميل
	pongWait = 60 * time.Second

	// كل قد إيه نبعث ping علشان نحافظ على الاتصال
	pingPeriod = (pongWait * 9) / 10

	// أقصى حجم للرسالة
	maxMessageSize = 512
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// اقرأ رسائل جاية من المتصفح وابعتها للـ hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		err := c.conn.Close()
		if err != nil {
			log.Println("Read Error : ", err)
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log.Println("SetReadDeadline : ", err)
	}
	c.conn.SetPongHandler(func(string) error {
		err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			log.Println("SetReadDeadline : ", err)
			return err
		}
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("ReadMessage : ", err)
			break
		}
		message = bytes.TrimSpace(message)
		c.hub.broadcast <- message
	}
}

// ابعت رسائل من hub للمتصفح
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.conn.Close()
		if err != nil {
			log.Println("Close Error : ", err)
			return
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Println("SetWriteDeadline Error : ", err)
				return
			}
			if !ok {
				// لو القناة اتقفلت
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Println("WriteMessage Error : ", err)
				}
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println("NextWriter Error : ", err)
				return
			}
			_, err = w.Write(message)
			if err != nil {
				log.Println("Write Error : ", err)
				return
			}

			// ابعت باقي الرسائل لو فيه
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, err := w.Write([]byte{'\n'})
				if err != nil {
					log.Println("Write Error 1 : ", err)
					return
				}
				_, err = w.Write(<-c.send)
				if err != nil {
					log.Println("Write Error 2 : ", err)
					return
				}
			}
			err = w.Close()
			if err != nil {
				log.Println("Close Error : ", err)
				return
			}

		case <-ticker.C:
			// ping للمتصفح علشان نحافظ على الاتصال
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Println("SetWriteDeadline Error : ", err)
				return
			}
			err = c.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Println("WriteMessage Error : ", err)
				return
			}
		}
	}
}
