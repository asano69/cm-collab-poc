package collab

import (
	"encoding/json"
	"log"

	"github.com/asano69/myapp/internal/lineid"
)

// lineInfo is what a client is told about one line.
type lineInfo struct {
	ID      string `json:"id"`
	Updated int64  `json:"updated"` // Unix time of the last edit that touched the line
	Admin   string `json:"admin"`   // administrator of the server the edit came through
}

// linesMessage is sent by the server: the info of every line, in line order.
type linesMessage struct {
	Type  string     `json:"type"`
	Lines []lineInfo `json:"lines"`
}

func encodeLines(meta lineid.Meta) []byte {
	lines := make([]lineInfo, len(meta.Lines))
	for i, line := range meta.Lines {
		lines[i] = lineInfo{ID: line.ID, Updated: line.Updated, Admin: line.Admin}
	}
	data, err := json.Marshal(linesMessage{Type: "lines", Lines: lines})
	if err != nil {
		// Unreachable: the message only holds strings and numbers.
		log.Printf("collab: encode lines: %v", err)
		return nil
	}
	return data
}
