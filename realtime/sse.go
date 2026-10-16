package realtime

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ishubhamsingh2e/bourbon/core"
)

type sseClient struct {
	id      string
	msgChan chan Message
	closed  bool
}

func newSSEClient() *sseClient {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return &sseClient{
		id:      hex.EncodeToString(b),
		msgChan: make(chan Message, 64),
	}
}

func (c *sseClient) ID() string {
	return c.id
}

func (c *sseClient) Send(msg Message) error {
	select {
	case c.msgChan <- msg:
		return nil
	default:
		return nil // Drop message if client buffer is backed up
	}
}

func (c *sseClient) Close() error {
	if !c.closed {
		c.closed = true
		close(c.msgChan)
	}
	return nil
}

// SSEHandler returns a Bourbon core.HandlerFunc that streams real-time Server-Sent Events from the Hub.
func SSEHandler(hub Hub, channel string, heartbeatInterval ...time.Duration) core.HandlerFunc {
	hbDuration := 15 * time.Second
	if len(heartbeatInterval) > 0 && heartbeatInterval[0] > 0 {
		hbDuration = heartbeatInterval[0]
	}

	return func(c *core.Context) error {
		flusher, ok := c.Response.(http.Flusher)
		if !ok {
			return fmt.Errorf("realtime: streaming unsupported, response writer does not implement http.Flusher")
		}

		c.Response.Header().Set("Content-Type", "text/event-stream")
		c.Response.Header().Set("Cache-Control", "no-cache")
		c.Response.Header().Set("Connection", "keep-alive")
		c.Response.Header().Set("X-Accel-Buffering", "no")
		c.Response.WriteHeader(http.StatusOK)
		flusher.Flush()

		client := newSSEClient()
		unsub, err := hub.Subscribe(c.Ctx, channel, client)
		if err != nil {
			return err
		}
		defer unsub()
		defer client.Close()

		ticker := time.NewTicker(hbDuration)
		defer ticker.Stop()

		for {
			select {
			case <-c.Request.Context().Done():
				return nil
			case <-ticker.C:
				_, _ = fmt.Fprint(c.Response, ": keep-alive\n\n")
				flusher.Flush()
			case msg, ok := <-client.msgChan:
				if !ok {
					return nil
				}
				dataBytes, err := json.Marshal(msg.Data)
				if err != nil {
					continue
				}

				if msg.Event != "" {
					_, _ = fmt.Fprintf(c.Response, "event: %s\n", msg.Event)
				}
				_, _ = fmt.Fprintf(c.Response, "data: %s\n\n", string(dataBytes))
				flusher.Flush()
			}
		}
	}
}
