// Package collab is the authority side of @codemirror/collab: it keeps the
// ordered list of accepted updates and relays them to every connected client.
package collab

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

// Update is one client's change set. Changes is kept opaque: the server never
// applies it, it only orders and relays it.
type Update struct {
	ClientID string          `json:"clientID"`
	Changes  json.RawMessage `json:"changes"`
}

// pushMessage is sent by a client: its synced version plus unconfirmed updates.
type pushMessage struct {
	Version int      `json:"version"`
	Updates []Update `json:"updates"`
}

// updatesMessage is sent by the server: updates starting at index From.
type updatesMessage struct {
	From    int      `json:"from"`
	Updates []Update `json:"updates"`
}

type client struct {
	send   chan []byte
	cancel context.CancelFunc
}

// Hub holds the single shared document.
type Hub struct {
	mu      sync.Mutex
	updates []Update
	clients map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{updates: []Update{}, clients: map[*client]struct{}{}}
}

func encode(from int, updates []Update) []byte {
	data, err := json.Marshal(updatesMessage{From: from, Updates: updates})
	if err != nil {
		// Unreachable: every Update was already parsed from valid JSON.
		log.Printf("collab: encode: %v", err)
		return nil
	}
	return data
}

// deliver queues data for c. A client that cannot keep up is dropped; it can
// reconnect and replay the whole history.
func deliver(c *client, data []byte) {
	select {
	case c.send <- data:
	default:
		c.cancel()
	}
}

func (h *Hub) join(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
	if len(h.updates) > 0 {
		deliver(c, encode(0, h.updates))
	}
}

func (h *Hub) leave(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
}

func validate(updates []Update) error {
	for _, u := range updates {
		if u.ClientID == "" || len(u.Changes) == 0 {
			return fmt.Errorf("update is missing clientID or changes")
		}
	}
	return nil
}

// push accepts the updates if the client is up to date, otherwise it sends the
// client what it is missing so that it can rebase and retry.
func (h *Hub) push(c *client, msg pushMessage) error {
	if err := validate(msg.Updates); err != nil {
		return err
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	switch {
	case msg.Version < 0 || msg.Version > len(h.updates):
		return fmt.Errorf("invalid version %d (server has %d updates)", msg.Version, len(h.updates))
	case msg.Version < len(h.updates):
		deliver(c, encode(msg.Version, h.updates[msg.Version:]))
	case len(msg.Updates) > 0:
		h.updates = append(h.updates, msg.Updates...)
		data := encode(msg.Version, msg.Updates)
		for cl := range h.clients {
			deliver(cl, data)
		}
	}
	return nil
}

func (c *client) writeLoop(ctx context.Context, conn *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-c.send:
			if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
				c.cancel()
				return
			}
		}
	}
}

func (h *Hub) readLoop(ctx context.Context, conn *websocket.Conn, c *client) error {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return err
		}
		var msg pushMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return fmt.Errorf("invalid message: %w", err)
		}
		if err := h.push(c, msg); err != nil {
			return err
		}
	}
}

// ServeHTTP upgrades the request to a WebSocket and serves one client.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("collab: accept: %v", err)
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(1 << 20)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	c := &client{send: make(chan []byte, 64), cancel: cancel}
	h.join(c)
	defer h.leave(c)
	go c.writeLoop(ctx, conn)

	if err := h.readLoop(ctx, conn, c); err != nil {
		// The reason is shown to the user by the frontend.
		conn.Close(websocket.StatusPolicyViolation, err.Error())
	}
}
