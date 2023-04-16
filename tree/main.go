package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"tree/messages/servermessages"
	"tree/treemanager"

	wskeyauth "github.com/clubcabana/ws-key-auth/go"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// TODO: this should really be its own thing.
//
//   The tree logic should its own package. Allowing us to easily swap out from
//   one implementation to the other.

var upgrader = websocket.Upgrader{}

var trees = treemanager.NewTreeManager[string, participant]()

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// TODO: Gotta find a better name for this.
//
// Perhaps move this to another file
type participant struct {
	conn *websocket.Conn
	meta json.RawMessage
}

var _ json.Marshaler = &participant{}

func (p *participant) MarshalJSON() ([]byte, error) {
	return p.meta, nil
}

type TypeData struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func handleTree(w http.ResponseWriter, r *http.Request) {
	// This is where we handle the act of adding a node to a tree

	params := mux.Vars(r)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	c.SetReadDeadline(time.Now().Add(pongWait))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	done := make(chan interface{})

	close := func() {
		done <- true
	}

	write := func(handler func() error) {
		err := c.SetWriteDeadline(time.Now().Add(writeWait))
		if err != nil {
			close()
		}
		err = handler()
		if err != nil {
			close()
		}
	}

	// TODO: handle pings and pings

	treeID, ok := params["id"]
	if !ok {
		// This should have technically not been possible at all. Thus closing the
		// connection, while also notifying the client that something went wrong.
		write(func() error {
			return c.WriteJSON(
				servermessages.CreateServerError(
					servermessages.ErrorResponse{Title: "An internal server error"},
				),
			)
		})
		return
	}

	ok, clientID, err := wskeyauth.Handshake(c)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return
	}
	if !ok {
		return
	}

	p := participant{c, json.RawMessage([]byte("{}"))}

	trees.Upsert(treeID, clientID, p)
	defer trees.DeleteNode(treeID, clientID)

	listener := trees.RegisterChangeListener(treeID)
	defer trees.UnregisterChangeListener(treeID, listener)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		defer func() { done <- true }()

		for {

			var td TypeData
			err := c.ReadJSON(&td)
			if err != nil {
				// Just kill the connection
				return
			}

			switch td.Type {
			case "SET_META":
				trees.Upsert(treeID, clientID, participant{c, td.Data})
			}
		}
	}()

	go func() {
		defer wg.Done()

		type typeAny struct {
			Type string `json:"type"`
			Data any    `json:"data"`
		}

		for {
			select {
			case <-listener:
				neighbors, ok := trees.GetNeighborOfNode(treeID, clientID)
				if ok {
					write(func() error {
						return c.WriteJSON(
							typeAny{
								Type: "NEIGHBORS",
								Data: neighbors,
							},
						)
					})
				}
			case <-done:
				return
			}
		}

	}()

	go func() {
		defer wg.Done()

		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	wg.Wait()
}

func handleWatchTree(w http.ResponseWriter, r *http.Request) {
	// This is where clients running diagnostics on a tree can peer into the state
	// of the tree

	params := mux.Vars(r)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	// TODO: handle pings and pongs

	treeId, ok := params["id"]
	if !ok {
		// This should have technically not been possible at all. Thus closing the
		// connection, while also notifying the client that something went wrong.
		c.WriteJSON(
			map[string]interface{}{
				"type": "SERVER_ERROR",
				"data": map[string]interface{}{
					"title": "An internal server error",
				},
			},
		)
		return
	}

	writeTree := func() error {

		return c.WriteJSON(
			map[string]interface{}{
				"type": "TREE",
				"data": trees.
					GetTree(treeId).
					AdjacencyList(),
			},
		)

	}

	writeTree()

	listener := trees.RegisterChangeListener(treeId)
	defer trees.UnregisterChangeListener(treeId, listener)

	for {
		<-listener
		if writeTree() != nil {
			return
		}
	}

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/tree/{id}", handleTree).Methods("UPGRADE")
	r.HandleFunc("/tree/{id}/watch", handleWatchTree).Methods("UPGRADE")
}
