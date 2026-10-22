package collab

import (
	"log"
	"time"

	"github.com/asano69/myapp/internal/lineid"
)

// serverID is the clientID of updates that the server makes itself. It never
// matches a real client, so every client treats them as remote updates.
const serverID = "server"

// Text returns the current document.
func (h *Hub) Text() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return fromUnits(h.doc)
}

// Replace makes text the document. The difference is recorded as an update
// from the server, so connected clients rebase their unconfirmed edits onto it.
func (h *Hub) Replace(text string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	changes, ok := replaceChanges(h.doc, toUnits(text))
	if !ok {
		return
	}
	updates := []Update{{ClientID: serverID, Changes: changes}}
	doc, meta, err := applyUpdates(h.doc, h.meta, updates, time.Now().Unix())
	if err != nil {
		// Unreachable: replaceChanges only returns changes that fit h.doc.
		log.Printf("collab: replace: %v", err)
		return
	}
	h.doc, h.meta = doc, meta
	h.accept(updates)
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
	h.mu.Lock()
	defer h.mu.Unlock()

	switch {
	case len(saved.Lines) == 0:
		// Nothing was saved yet.
	case len(saved.Lines) == lineid.Count(h.doc):
		h.meta = saved
	default:
		log.Printf("collab: saved line IDs cover %d lines but the document has %d; assigning new IDs",
			len(saved.Lines), lineid.Count(h.doc))
		h.meta = lineid.New(h.doc, time.Now().Unix())
	}
}
