package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Swunci/rss-feed-backend/internal/models"
	"github.com/Swunci/rss-feed-backend/internal/services"
	"github.com/go-chi/chi/v5"
)

type CollectionHandler struct {
	collectionService *services.CollectionService
}

func NewCollectionHandler(cs *services.CollectionService) *CollectionHandler {
	return &CollectionHandler{collectionService: cs}
}

func (h *CollectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	collections, err := h.collectionService.GetCollections()
	if err != nil {
		slog.Error("Fetch collections", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Debug("Collections fetched", "count", len(collections))
	WriteJSON(w, http.StatusOK, collections)
}

func (h *CollectionHandler) Post(w http.ResponseWriter, r *http.Request) {
	var req models.CollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	collection, err := h.collectionService.CreateCollection(req.Name)
	if err != nil {
		slog.Error("Create collection", "name", req.Name, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Debug("Collections created", "id", collection.ID, "name", collection.Name)
	WriteJSON(w, http.StatusOK, collection)
}

func (h *CollectionHandler) Put(w http.ResponseWriter, r *http.Request) {
	collection_id, err := ParseID(chi.URLParam(r, "collection_id"))
	if err != nil {
		http.Error(w, "invalid url parameter", http.StatusBadRequest)
		return
	}
	var req models.CollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	err = h.collectionService.UpdateCollection(collection_id, req.Name)
	if err != nil {
		slog.Error("Update collection", "collection_id", collection_id, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *CollectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	collection_id, err := ParseID(chi.URLParam(r, "collection_id"))
	if err != nil {
		http.Error(w, "invalid url parameter", http.StatusBadRequest)
		return
	}
	err = h.collectionService.DeleteCollection(collection_id)
	if err != nil {
		slog.Error("Delete collection", "collection_id", collection_id, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
