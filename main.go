package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/PhilippReinke/synced-canvas/data"
	"github.com/PhilippReinke/synced-canvas/wsm"
	"golang.org/x/net/websocket"
)

const (
	addr = ":8080"
)

func main() {
	// setup server, canvas and websocket manager
	canvas := data.NewCanvas()
	manager := wsm.NewManager(canvas.ProcessNewMessage)
	server := &http.Server{
		Addr: addr,
	}
	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.Handle("/ws", websocket.Handler(manager.HandleWS))
	http.HandleFunc("/canvas/lines", canvas.GetLinesHandler)
	http.HandleFunc("/canvas/reset", canvas.ResetHandler)

	// run HTTP server
	go func() {
		slog.Info("Server running.", "addr", addr)
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				// do not log failure on graceful shutdown
				slog.Error("Server failure.", "err", err)
			}
		}
	}()

	// block until interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	slog.Info("Blocking, press ctrl+c to terminate app.")
	<-stop

	// cleanup
	slog.Info("Shutting down HTTP server...")
	if err := server.Shutdown(context.Background()); err != nil {
		slog.Error("Graceful shutdown of HTTP server failed.", "err", err)
	}
	slog.Info("Closing websocket connections...")
	manager.CloseAllConns()

	// goodbye message
	slog.Info("Graceful shutdown complete :)")
}
