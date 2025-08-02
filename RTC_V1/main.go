package main

import (
	"log"
	"net/http"
)

func main() {
	liveHub := NewHub()
	go liveHub.Run()

	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		ServerWs(liveHub, writer, request)
	})

	log.Println("Starting server on port : 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Listen And Server : ", err)
	}
}
