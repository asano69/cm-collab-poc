// Package linehub is the collab server of this app. It adds the line IDs and
// their signatures to collab.Server: every line of the document has an ID and
// is signed with the author of its latest edit.
package linehub

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/asano69/cmgo/authority"
	"github.com/asano69/cmgo/collab"
	"github.com/asano69/cmgo/ot"
	"github.com/asano69/myapp/internal/lineid"
	"github.com/asano69/myapp/internal/relayauth"
)

// relayAdmin is the administrator of the relay server, a fixed name in this
// proof of concept. The main server's is serverAuthor.
const relayAdmin = "Bob"

// serverAuthor is who edits made by the main server itself, and edits of
// directly connected browsers, are attributed to. The administrator's name is
// fixed in this proof of concept.
var serverAuthor = lineid.Author{Admin: "Alice"}

// Signer signs the lines of a document that were changed by author at time now.
type Signer interface {
	Sign(doc []uint16, meta lineid.Meta, author lineid.Author, now int64) lineid.Meta
}

// Hub is a collab.Server that keeps the line IDs of its document.
type Hub struct {
	*collab.Server

	// mu guards meta and signer. Lock order: the Server's lock, then mu.
	mu     sync.Mutex
	meta   lineid.Meta // the line IDs of the document, changed together with it
	signer Signer      // signs the changed lines of meta; nil means no signing
}

func New() *Hub {
	h := &Hub{meta: lineid.New(nil, time.Now().Unix())}
	h.Server = collab.New(collab.WithTracker(h), collab.WithAuthor(authorOf))
	return h
}

// authorOf names who is behind a connection: a relay that signed its request
// with a key, or a browser that connected directly.
func authorOf(r *http.Request) any {
	if key, ok := relayauth.KeyFrom(r.Context()); ok {
		return lineid.Author{Admin: relayAdmin, Key: relayauth.PublicKeyLine(key)}
	}
	return serverAuthor
}

// Track implements collab.Tracker. A nil author (an edit made by the server
// itself) is attributed to serverAuthor.
func (h *Hub) Track(changes []authority.Change, doc []uint16, author any, now int64) {
	a, ok := author.(lineid.Author)
	if !ok {
		a = serverAuthor
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, c := range changes {
		h.meta = h.meta.Apply(c.Before, toEdits(c.Changes.Edits()), now)
	}
	h.meta = h.sign(doc, h.meta, a, now)
}

// Message implements collab.Tracker: the line IDs, sent to clients after the
// text that they belong to.
func (h *Hub) Message() []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	return encodeLines(h.meta)
}

func toEdits(edits []ot.Edit) []lineid.Edit {
	out := make([]lineid.Edit, len(edits))
	for i, e := range edits {
		out[i] = lineid.Edit{From: e.From, To: e.To, Text: e.Text}
	}
	return out
}

// sign signs the changed lines of meta, if the hub has a signer. h.mu must be
// held.
func (h *Hub) sign(doc []uint16, meta lineid.Meta, author lineid.Author, now int64) lineid.Meta {
	if h.signer == nil {
		return meta
	}
	return h.signer.Sign(doc, meta, author, now)
}

// SetSigner makes the hub sign the line IDs from now on, and signs the lines
// that exist already.
func (h *Hub) SetSigner(s Signer) {
	h.Inspect(func(doc []uint16) {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.signer = s
		h.meta = h.sign(doc, h.meta, serverAuthor, time.Now().Unix())
	})
}

// Meta returns the line IDs of the current document. The result is never
// modified afterwards, so it can be kept.
func (h *Hub) Meta() lineid.Meta {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.meta
}

// RestoreMeta adopts line IDs that were saved earlier, if they fit the
// document. If they do not (the file was edited while the server was down),
// every line gets a new ID, so an ID that was handed out before is never reused.
func (h *Hub) RestoreMeta(saved lineid.Meta) {
	h.Inspect(func(doc []uint16) {
		h.mu.Lock()
		defer h.mu.Unlock()

		switch {
		case len(saved.Lines) == 0:
			// Nothing was saved yet.
		case len(saved.Lines) == lineid.Count(doc):
			h.meta = saved
		default:
			log.Printf("linehub: saved line IDs cover %d lines but the document has %d; assigning new IDs",
				len(saved.Lines), lineid.Count(doc))
			h.meta = lineid.New(doc, time.Now().Unix())
		}
		// Lines that were edited while the server was down no longer match
		// their hash, and get signed again.
		h.meta = h.sign(doc, h.meta, serverAuthor, time.Now().Unix())
	})
}
