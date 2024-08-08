// Package wsm implements a Websocket Manager
package wsm

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"golang.org/x/net/websocket"
)

type MsgProcessor func([]byte, *Manager)

type Manager struct {
	mu    sync.RWMutex
	conns map[*websocket.Conn]context.CancelFunc
	// upstream messages are forwarded to callback
	msgProcessor MsgProcessor
}

func NewManager(msgProcessor MsgProcessor) *Manager {
	return &Manager{
		conns:        make(map[*websocket.Conn]context.CancelFunc),
		msgProcessor: msgProcessor,
	}
}

func (m *Manager) AddConn(ws *websocket.Conn, cancel context.CancelFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[ws] = cancel
}

func (m *Manager) CloseConn(ws *websocket.Conn) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cancel, ok := m.conns[ws]
	if !ok {
		return nil
	}

	cancel() // cancel read loop
	err := ws.Close()
	delete(m.conns, ws)
	if err != nil {
		return fmt.Errorf("could not close ws connection: %w", err)
	}
	return nil
}

func (m *Manager) CloseAllConns() {
	m.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(m.conns))
	for ws := range m.conns {
		conns = append(conns, ws)
	}
	m.mu.RUnlock()

	for _, ws := range conns {
		if err := m.CloseConn(ws); err != nil {
			slog.Error("Failed to close connection.",
				"err", err,
				"addr", ws.RemoteAddr(),
			)
		}
	}
}

func (m *Manager) ConnCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.conns)
}

func (m *Manager) Broadcast(msg []byte) {
	m.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(m.conns))
	for ws := range m.conns {
		conns = append(conns, ws)
	}
	m.mu.RUnlock()

	for _, ws := range conns {
		_, err := ws.Write(msg)
		if err != nil {
			slog.Error("Failed to broadcast message.",
				"err", err,
				"addr", ws.RemoteAddr(),
			)
			m.CloseConn(ws)
		}
	}
}

func (m *Manager) HandleWS(ws *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	m.AddConn(ws, cancel)
	m.readLoop(ctx, ws)
}

func (m *Manager) readLoop(ctx context.Context, ws *websocket.Conn) {
	defer m.CloseConn(ws)
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF || ctx.Err() != nil {
				break
			}
			slog.Error("Websocket read error.",
				"err", err,
				"addr", ws.RemoteAddr(),
			)
			return
		}
		m.msgProcessor(buf[:n], m)
	}
}
