package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
)

type ItemSSEHandler struct {
	itemSEEChannel chan string
}

func NewItemSSEHandler(itemSEEChannel chan string) *ItemSSEHandler {
	return &ItemSSEHandler{itemSEEChannel: itemSEEChannel}
}

func (h *ItemSSEHandler) ItemEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}
	for {
		select {
		case msg := <-h.itemSEEChannel:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			slog.Debug("SSE event sent", "msg", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
