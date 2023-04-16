package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"tree/messages/servermessages"
	"tree/treemanager"

	wskeyauth "github.com/clubcabana/ws-key-auth/go"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/shovon/gorillawswrapper"
)

var upgrader = websocket.Upgrader{}

var trees = treemanager.NewTreeManager[string, participant]()

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

func handleTree(w http.ResponseWriter, r *http.Request) {
	// This is where we handle the act of adding a node to a tree

	params := mux.Vars(r)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	treeId, ok := params["id"]
	if !ok {
		// This should have technically not been possible at all. Thus closing the
		// connection, while also notifying the client that something went wrong.
		c.WriteJSON(servermessages.CreateServerError(servermessages.ErrorResponse{Title: "An internal server error"}))
		return
	}

	clientId := r.URL.Query().Get("client_id")
	if len(clientId) <= 0 {
		// This is entirely possible. So if we're here, then notify the client that
		// they made a bad request (althoug, to be fair, it *could* also be because
		// the backend was coded poorly. This needs to be accounted for)
		c.WriteJSON(
			servermessages.CreateClientError(
				servermessages.ErrorResponse{Title: "A client ID was not supplied"},
			),
		)
		return
	}

	{
		ok, err := wskeyauth.Handshake(c)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			return
		}
		if !ok {
			return
		}
	}

	p := participant{c, json.RawMessage([]byte("{}"))}

	trees.Upsert(treeId, clientId, p)
	defer trees.DeleteNode(treeId, clientId)

	listener := trees.RegisterChangeListener(treeId)
	defer trees.UnregisterChangeListener(treeId, listener)

	var wg sync.WaitGroup
	done := make(chan interface{})
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer func() { done <- true }()

		for {
			type typeData struct {
				Type string          `json:"type"`
				Data json.RawMessage `json:"data"`
			}
			var td typeData
			err := c.ReadJSON(&td)
			if err != nil {
				// Maybe the connection has closed?
				return
			}
		}
	}()

	go func() {
		defer wg.Done()

		select {
		case <-listener:
		case <-done:
			return
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

	treeId, ok := params["id"]
	if !ok {
		// This should have technically not been possible at all. Thus closing the
		// connection, while also notifying the client that something went wrong.
		c.WriteJSON(
			servermessages.
				CreateServerError(
					servermessages.ErrorResponse{Title: "An internal server error"},
				),
		)
		return
	}

	conn := gorillawswrapper.NewWrapper(c)

	conn.WriteJSON(
		servermessages.
			CreateWholeTreeMessage(
				trees.
					GetTree(treeId).
					AdjacencyList(),
			),
	)

	mc := conn.MessagesChannel()
	listener := trees.RegisterChangeListener(treeId)
	defer trees.UnregisterChangeListener(treeId, listener)

	for {
		select {
		case _, ok := <-mc:
			if !ok {
				return
			}
		case <-listener:
			// TODO: send the entire tree to the client that is watching the tree
		}
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/tree/{id}", handleTree).Methods("UPGRADE")
	r.HandleFunc("/tree/{id}/watch", handleWatchTree).Methods("UPGRADE")
}
