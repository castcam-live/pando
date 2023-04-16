package ws

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10
)

type Writer struct {
	conn *websocket.Conn
}

func NewWriter(conn *websocket.Conn) Writer {
	return Writer{conn: conn}
}

func (c Writer) writeDeadLine() error {
	return c.conn.SetWriteDeadline(time.Now().Add(WriteWait))

}

func (c Writer) WriteJSON(v any) error {
	return c.conn.WriteJSON(v)
}

func (c Writer) WriteMessage(messageType int, data []byte) error {
	err := c.writeDeadLine()
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(messageType, data)
}
