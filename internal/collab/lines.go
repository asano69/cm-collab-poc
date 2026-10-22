package collab

import (
	"encoding/json"
	"log"

	"github.com/asano69/myapp/internal/lineid"
)

// linesMessage is sent by the server: the ID of every line, in line order.
type linesMessage struct {
	Type string  `json:"type"`
	IDs  []int64 `json:"ids"`
}

func encodeLines(meta lineid.Meta) []byte {
	ids := make([]int64, len(meta.Lines))
	for i, line := range meta.Lines {
		ids[i] = line.ID
	}
	data, err := json.Marshal(linesMessage{Type: "lines", IDs: ids})
	if err != nil {
		// Unreachable: the message only holds a string and numbers.
		log.Printf("collab: encode lines: %v", err)
		return nil
	}
	return data
}
