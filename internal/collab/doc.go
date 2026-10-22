package collab

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

	next := toUnits(text)
	changes, ok := replaceChanges(h.doc, next)
	if !ok {
		return
	}
	h.doc = next
	h.accept([]Update{{ClientID: serverID, Changes: changes}})
}
