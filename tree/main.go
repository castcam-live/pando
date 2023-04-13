package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"tree/messages/servermessages"
	"tree/treemanager"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/shovon/gorillawswrapper"
	"github.com/sparkscience/wskeyid-go/v2"
)

var upgrader = websocket.Upgrader{}

var trees = treemanager.NewTreeManager[string, participant]()

// TODO: Gotta find a better name for this.
//
// Perhaps move this to another file
type participant struct {
	conn gorillawswrapper.Wrapper
	meta json.RawMessage
}

var _ json.Marshaler = &participant{}

func (p *participant) MarshalJSON() ([]byte, error) {
	return p.meta, nil
}

func handleTree(w http.ResponseWriter, r *http.Request) {
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

	conn := gorillawswrapper.NewWrapper(c)

	{
		err := wskeyid.HandleAuthConnection(r, conn)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			return
		}
	}

	p := participant{conn, json.RawMessage([]byte("{}"))}

	trees.Upsert(treeId, clientId, p)
	defer trees.DeleteNode(treeId, clientId)

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
			// TODO: emit something to the client
		}
	}

	// The end
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
