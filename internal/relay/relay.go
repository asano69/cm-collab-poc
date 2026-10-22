// Package relay forwards editor connections to the main server. Each browser
// connection gets its own signed connection upstream, and messages are copied
// in both directions without being looked at, so the main server sees every
// browser as an ordinary client.
package relay

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coder/websocket"

	"github.com/asano69/myapp/internal/relayauth"
)

const (
	// readLimit bounds what a browser may send, as on the main server.
	readLimit = 1 << 20
	// maxReason is the longest close reason that a WebSocket can carry.
	maxReason = 123
)

// Proxy is an http.Handler that relays WebSocket connections.
type Proxy struct {
	upstream *url.URL
	key      ed25519.PrivateKey
}

// New returns a Proxy for the main server at upstream (ws:// or wss://).
func New(upstream string, key ed25519.PrivateKey) (*Proxy, error) {
	u, err := url.Parse(strings.TrimRight(upstream, "/"))
	if err != nil || (u.Scheme != "ws" && u.Scheme != "wss") || u.Host == "" {
		return nil, fmt.Errorf("upstream %q must look like ws://host:port or wss://host:port", upstream)
	}
	u.Path += relayauth.Path
	return &Proxy{upstream: u, key: key}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	client, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("relay: accept: %v", err)
		return
	}
	defer client.CloseNow()
	client.SetReadLimit(readLimit)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	upstream, err := p.dial(ctx)
	if err != nil {
		// The browser shows the reason to the user.
		log.Printf("relay: %v", err)
		client.Close(websocket.StatusPolicyViolation, truncate(err.Error()))
		return
	}
	defer upstream.CloseNow()
	// The main server is trusted, and replaying the history can be large.
	upstream.SetReadLimit(-1)

	go func() {
		pipe(ctx, upstream, client)
		cancel()
	}()
	pipe(ctx, client, upstream)
	cancel()
}

// dial opens the signed connection to the main server. The error text is
// meant for the user.
func (p *Proxy) dial(ctx context.Context) (*websocket.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	header := http.Header{}
	relayauth.Sign(header, p.key, p.upstream.Path, time.Now())
	conn, resp, err := websocket.Dial(ctx, p.upstream.String(), &websocket.DialOptions{HTTPHeader: header})
	if err == nil {
		return conn, nil
	}
	if resp != nil && resp.Body != nil {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		if msg := strings.TrimSpace(string(body)); msg != "" {
			return nil, errors.New("main server refused the relay: " + msg)
		}
	}
	return nil, fmt.Errorf("cannot reach the main server: %w", err)
}

// pipe copies messages from src to dst until one of them fails. When src was
// closed, dst is closed with the same status and reason.
func pipe(ctx context.Context, dst, src *websocket.Conn) {
	for {
		typ, data, err := src.Read(ctx)
		if err != nil {
			closeLike(dst, err)
			return
		}
		if err := dst.Write(ctx, typ, data); err != nil {
			return
		}
	}
}

// closeLike closes c with the status and reason of the close error err.
func closeLike(c *websocket.Conn, err error) {
	var ce websocket.CloseError
	if !errors.As(err, &ce) {
		c.CloseNow()
		return
	}
	if c.Close(ce.Code, truncate(ce.Reason)) != nil {
		c.CloseNow()
	}
}

func truncate(reason string) string {
	if len(reason) > maxReason {
		return reason[:maxReason]
	}
	return reason
}
