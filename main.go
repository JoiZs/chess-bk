package main

import (
	"context"
	"net/http"

	"github.com/JoiZs/chess-bk/initi"
	"github.com/JoiZs/chess-bk/ws"
)

func main() {
	// initialize the project
	initi.InitProj()

	ctx := context.Background()

	wsM := ws.InitManager(ctx)

	// http connection
	http.HandleFunc("/ws", wsM.WsHandler)

	http.ListenAndServe(":4444", nil)
}
