package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// IPCClient connects to the daemon socket and receives broadcast messages.
type IPCClient struct {
	socketPath string
	conn       net.Conn
	enc        *json.Encoder
}

// Dial connects to the daemon at socketPath with a short timeout.
func Dial(socketPath string) (*IPCClient, error) {
	conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("ipc: dial %s: %w", socketPath, err)
	}
	return &IPCClient{
		socketPath: socketPath,
		conn:       conn,
		enc:        json.NewEncoder(conn),
	}, nil
}

// Close closes the connection.
func (c *IPCClient) Close() error { return c.conn.Close() }

// Send sends a typed message to the daemon.
func (c *IPCClient) Send(msgType string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.enc.Encode(Envelope{Type: msgType, Payload: raw})
}

// Receive reads messages from the daemon and passes each to the callback.
// Blocks until the connection is closed.
func (c *IPCClient) Receive(fn func(Envelope)) error {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var env Envelope
		if err := json.Unmarshal(line, &env); err != nil {
			continue
		}
		fn(env)
	}
	return scanner.Err()
}
