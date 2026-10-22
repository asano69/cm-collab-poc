package relay

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"net/http/httptest"

	"github.com/coder/websocket"

	"github.com/asano69/myapp/internal/linehub"
	"github.com/asano69/myapp/internal/relayauth"
)

func wsURL(httpURL string) string { return "ws" + strings.TrimPrefix(httpURL, "http") }

// startRelay starts a main server that authorizes the keys in authorized, and
// a relay that signs with key. It returns the relay's URL and the hub.
func startRelay(t *testing.T, authorized string) (string, *linehub.Hub) {
	t.Helper()
	dir := t.TempDir()
	priv, err := relayauth.GenerateKey(filepath.Join(dir, "key.pem"))
	if err != nil {
		t.Fatal(err)
	}
	line := relayauth.AuthorizedKey(priv)
	if authorized == "other" {
		other, err := relayauth.GenerateKey(filepath.Join(dir, "other.pem"))
		if err != nil {
			t.Fatal(err)
		}
		line = relayauth.AuthorizedKey(other)
	}
	keys := filepath.Join(dir, "authorized_keys")
	if err := os.WriteFile(keys, []byte(line+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	hub := linehub.New()
	main := httptest.NewServer(relayauth.Require(keys, hub))
	t.Cleanup(main.Close)

	proxy, err := New(wsURL(main.URL), priv)
	if err != nil {
		t.Fatal(err)
	}
	front := httptest.NewServer(proxy)
	t.Cleanup(front.Close)
	return wsURL(front.URL), hub
}

func TestProxyForwardsEdits(t *testing.T) {
	front, hub := startRelay(t, "own")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, front, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	push := `{"type":"push","version":0,"updates":[{"clientID":"a","changes":[[0,"hi"]]}]}`
	if err := c.Write(ctx, websocket.MessageText, []byte(push)); err != nil {
		t.Fatal(err)
	}
	// Every client first gets the line IDs, then the reply to its push.
	for _, want := range []string{`"type":"lines"`, `"type":"updates"`} {
		_, data, err := c.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), want) {
			t.Fatalf("reply = %s, want a message with %s", data, want)
		}
	}
	if got := hub.Text(); got != "hi" {
		t.Errorf("hub text = %q, want %q", got, "hi")
	}
}

func TestProxyShowsWhyTheMainServerRefused(t *testing.T) {
	front, _ := startRelay(t, "other")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, front, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	_, _, err = c.Read(ctx)
	var ce websocket.CloseError
	if !errors.As(err, &ce) {
		t.Fatalf("error = %v, want a close error", err)
	}
	if !strings.Contains(ce.Reason, "not in authorized_keys") {
		t.Errorf("close reason = %q, want it to explain the refusal", ce.Reason)
	}
}

func TestNewRejectsBadUpstream(t *testing.T) {
	for _, u := range []string{"", "http://host:1", "host:1"} {
		if _, err := New(u, nil); err == nil {
			t.Errorf("New(%q) succeeded, want an error", u)
		}
	}
}
