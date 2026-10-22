// Package metasync saves the line IDs of a document to a file.
package metasync

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/asano69/cmgo/filesync"
	"github.com/asano69/myapp/internal/lineid"
)

// saveInterval is how often changed line IDs are written to the file.
const saveInterval = time.Second

// MetaSource holds the line IDs of a document.
type MetaSource interface {
	Meta() lineid.Meta
	RestoreMeta(lineid.Meta)
}

// Start restores the line IDs from path, then saves them there every
// second when they changed. The file is read only once: the server is its only
// writer. Call it after filesync.Start, so that the text is loaded when the IDs
// are matched against it.
func Start(src MetaSource, path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve %q: %w", path, err)
	}
	saved, err := loadMeta(abs)
	if err != nil {
		return err
	}
	src.RestoreMeta(saved)

	m := &metaSaver{src: src, path: abs}
	if err := m.save(); err != nil {
		return err
	}
	go m.run()
	return nil
}

func loadMeta(path string) (lineid.Meta, error) {
	var meta lineid.Meta
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return meta, nil
	}
	if err != nil {
		return meta, fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return meta, fmt.Errorf("%s is not valid (remove it to assign new line IDs): %w", path, err)
	}
	return meta, nil
}

type metaSaver struct {
	src  MetaSource
	path string
	// saved is what this process last wrote to the file.
	saved []byte
}

func (m *metaSaver) run() {
	for range time.Tick(saveInterval) {
		if err := m.save(); err != nil {
			log.Printf("filesync: %v", err)
		}
	}
}

func (m *metaSaver) save() error {
	data, err := json.Marshal(m.src.Meta())
	if err != nil {
		return fmt.Errorf("encode line IDs: %w", err)
	}
	if bytes.Equal(data, m.saved) {
		return nil
	}
	if err := filesync.WriteAtomic(m.path, string(data)); err != nil {
		return err
	}
	m.saved = data
	return nil
}
