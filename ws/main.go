package ws

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type WsConn struct {
	upgrader websocket.Upgrader
	clients  []websocket.Conn
}

func Wsocket() *WsConn {
	return &WsConn{
		upgrader: websocket.Upgrader{},
	}
}

func (wsc *WsConn) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsc.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	wsc.readLoop(conn)
}

func (wsc *WsConn) readLoop(connt *websocket.Conn) {
	for {
		messageType, data, err := connt.ReadMessage()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("Received message: %s\n", data)

		if err := connt.WriteMessage(messageType, data); err != nil {
			fmt.Println(err)
			return
		}
	}
}
