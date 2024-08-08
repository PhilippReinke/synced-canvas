package data

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/PhilippReinke/synced-canvas/wsm"
)

// CanvasMessage is used for en-/decoding json string messages for communication
// between client and server.
//
// Type represents the canvas element type for which the data are meant for.
// Only the line type has been implemented so far.
type CanvasMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Line struct {
	Points    []Point `json:"points"`
	Color     string  `json:"color"`
	LineWidth int     `json:"lineWidth"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Canvas handles storage of canvas data.
//
// TODO: persist data. now it is only in memory
type Canvas struct {
	mu    sync.RWMutex
	lines []Line
}

func NewCanvas() *Canvas {
	return &Canvas{}
}

func (c *Canvas) StoreLine(line Line) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lines = append(c.lines, line)
}

func (c *Canvas) ProcessNewMessage(msg []byte, manager *wsm.Manager) {
	canvasMessage := CanvasMessage{}
	if err := json.Unmarshal(msg, &canvasMessage); err != nil {
		slog.Error("Could not decode upstream message.",
			"err", err,
			"msg", msg,
		)
		return
	}
	if canvasMessage.Type != "line" {
		slog.Warn("Drpped upstream message with invalid type.", "msg", msg)
		return
	}
	line := Line{}
	if err := json.Unmarshal(canvasMessage.Data, &line); err != nil {
		slog.Error("Could not decode upstream message.",
			"err", err,
			"msg", msg,
		)
		return
	}
	c.StoreLine(line)
	manager.Broadcast(msg)
}

// GetLinesHandler implements an HTTP handler that allows to fetch all canvas
// lines in JSON format.
//
// TODO: One should enumerate elements and allow to fetch specific elements.
func (c *Canvas) GetLinesHandler(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	data := struct {
		Lines []Line `json:"lines"`
	}{
		Lines: c.lines,
	}
	c.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// ResetHandler deletes stored canvas data.
func (c *Canvas) ResetHandler(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	c.lines = []Line{}
	c.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}
