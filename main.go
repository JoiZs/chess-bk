package main

import (
	"log"
	"net/http"

	"github.com/JoiZs/chess-bk/initializer"
	"github.com/JoiZs/chess-bk/ws"
)

func main() {
	initializer.Init()

	wsServer := ws.Wsocket()

	http.HandleFunc("/ws", wsServer.WsHandler)

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
