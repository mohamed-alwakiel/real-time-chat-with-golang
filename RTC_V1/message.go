package main

type Message struct {
	Type    string `json:"type"` // join, leave, chat
	Sender  string `json:"sender"`
	Content string `json:"message"`
}
