package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"tree/graph/treemanager"
	"tree/ws"

	wskeyauth "github.com/clubcabana/ws-key-auth/go"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// TODO: this should really be its own thing.
//
//   The tree logic should be its own package. Allowing us to easily swap out
//   from one implementation to the other.

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var trees = treemanager.NewTreeManager[string, participant]()

// TODO: Gotta find a better name for this.
//
// Perhaps move this to another file
type participant struct {
	writer ws.Writer
	meta   json.RawMessage
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

	c.SetReadDeadline(time.Now().Add(ws.PongWait))
	c.SetPongHandler(func(string) error {
		c.SetReadDeadline(time.Now().Add(ws.PongWait))
		return nil
	})

	done := make(chan interface{})

	close := func() {
		done <- true
	}

	write := func(handler func() error) {
		err := c.SetWriteDeadline(time.Now().Add(ws.WriteWait))
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
				map[string]any{
					"type": "SERVER_ERROR",
					"data": map[string]any{
						"type": "UNKNOWN_ERROR",
						"data": map[string]string{
							"title": "An internal server error",
						},
					},
				},
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

	writer := ws.NewWriter(c)

	p := participant{writer, json.RawMessage([]byte("{}"))}

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
				trees.Upsert(treeID, clientID, participant{writer, td.Data})
			case "BROADCAST":
				neighbors, ok := trees.GetNeighborOfNode(treeID, clientID)

				if ok {
					for _, n := range neighbors {
						if n.Value.writer.WriteJSON(td.Data) != nil {
							writer.WriteJSON(map[string]any{
								"type": "SERVER_ERROR",
								"data": map[string]any{
									"type": "UNABLE_TO_SEND_MESSAGE",
									"data": map[string]any{
										"message": fmt.Sprintf("In broadcast, error sending message to participant with client ID of %s. Could be that the participant is no longer there", n.Key),
										"meta": map[string]any{
											"error":            err.Error(),
											"to":               n.Key,
											"original_message": td.Data,
										},
									},
								},
							})
						}
					}
				}
			case "SEND":
				type message struct {
					To   string          `json:"to"`
					Data json.RawMessage `json:"data"`
				}

				var m message
				err := json.Unmarshal(td.Data, &m)
				if err != nil {
					writer.WriteJSON(map[string]any{
						"type": "CLIENT_ERROR",
						"data": map[string]any{
							"type": "MALFORMED_MESSAGE",
							"data": map[string]any{
								"title":   "Message intended for participant was malformed",
								"message": fmt.Sprintf("Error parsing the message that was intended for participant %s", m.To),
								"meta": map[string]any{
									"error": err.Error(),
									"to":    m.To,
								},
							},
						},
					})
					continue
				}

				neighbors, ok := trees.GetNeighborOfNode(treeID, clientID)

				if ok {
					sent := false
					for _, n := range neighbors {
						if n.Key == m.To {
							sent = true
							if n.Value.writer.WriteJSON(m.Data) != nil {
								writer.WriteJSON(map[string]any{
									"type": "SERVER_ERROR",
									"data": map[string]any{
										"type": "UNABLE_TO_SEND_MESSAGE",
										"data": map[string]any{
											"message": fmt.Sprintf("Error sending message to participant with client ID of %s. Could be that the participant is no longer there", n.Key),
											"meta": map[string]any{
												"error":            err.Error(),
												"to":               m.To,
												"original_message": td.Data,
											},
										},
									},
								})
							}
							continue
						}
					}

					if !sent {
						writer.WriteJSON(map[string]any{
							"type": "CLIENT_ERROR",
							"data": map[string]any{
								"type": "PARTICIPANT_NOT_FOUND",
								"data": map[string]any{
									"message": fmt.Sprintf("Participant with ID %s not found", m.To),
									"meta": map[string]any{
										"to":               m.To,
										"original_message": td.Data,
									},
								},
							},
						})
					}
				}
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

		ticker := time.NewTicker(ws.PingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.SetWriteDeadline(time.Now().Add(ws.WriteWait))
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
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", GetPort()))
	if err != nil {
		panic(err)
	}
	fmt.Println("Listening on port", listener.Addr().(*net.TCPAddr).Port)
	panic(http.Serve(listener, r))
}
