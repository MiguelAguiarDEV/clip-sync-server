package app

import (
	"net/http"

	"clip-sync/server/internal/hub"
	"clip-sync/server/internal/ws"
)

func NewMux() http.Handler {
	mux := http.NewServeMux()

	// Healthcheck
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Hub + WebSocket /ws
	h := hub.New(32)
	wss := &ws.Server{
		Hub: h,
		Auth: func(token string) (string, bool) {
			if token == "" {
				return "", false
			}
			// MVP: token == userID
			return token, true
		},
	}
	mux.Handle("/ws", wss)

	return mux
}
