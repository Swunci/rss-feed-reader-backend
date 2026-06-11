package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Swunci/rss-feed-backend/internal/models"
	"github.com/Swunci/rss-feed-backend/internal/services"
	"github.com/go-chi/chi/v5"
)

type ItemHandler struct {
	itemService *services.ItemService
}

func NewItemHandler(s *services.ItemService) *ItemHandler {
	return &ItemHandler{itemService: s}
}

func (h *ItemHandler) Get(w http.ResponseWriter, r *http.Request) {
	item_id, err := ParseID(chi.URLParam(r, "item_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		println(err.Error())
		return
	}
	item, err := h.itemService.GetItem(item_id)
	if err != nil {
		slog.Error("Fetch item", "item_id", item_id, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *ItemHandler) GetAllItems(w http.ResponseWriter, r *http.Request) {
	timestamp_cursor := r.URL.Query().Get("cursor")
	limit := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	filter := parseItemFilter(r)
	slog.Debug("Item filters", "read", filter.IsRead, "favorite", filter.IsFavorite)

	items, err := h.itemService.GetAllItems(filter, timestamp_cursor, limit)
	if err != nil {
		slog.Error("Get all items", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Debug("Items fetched", "count", len(items))
	WriteJSON(w, http.StatusOK, items)
}

func (h *ItemHandler) GetItemsByFeed(w http.ResponseWriter, r *http.Request) {
	feed_id, err := ParseID(chi.URLParam(r, "feed_id"))
	timestamp_cursor := r.URL.Query().Get("cursor")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	limit := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	filter := parseItemFilter(r)
	slog.Debug("Item filters", "feed_id", feed_id, "read", filter.IsRead, "favorite", filter.IsFavorite)

	items, err := h.itemService.GetItemsByFeed(feed_id, filter, timestamp_cursor, limit)

	if err != nil {
		slog.Error("Get items by feed", "feed_id", feed_id, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Debug("Items fetched", "feed_id", feed_id, "count", len(items))
	WriteJSON(w, http.StatusOK, items)
}

func (h *ItemHandler) GetItemsByCollection(w http.ResponseWriter, r *http.Request) {
	collection_id, err := ParseID(chi.URLParam(r, "collection_id"))
	timestamp_cursor := r.URL.Query().Get("cursor")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	limit := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		limit = parsed
	}
	filter := parseItemFilter(r)
	slog.Debug("Item filters", "collection_id", collection_id, "read", filter.IsRead, "favorite", filter.IsFavorite)

	items, err := h.itemService.GetItemsByCollection(collection_id, filter, timestamp_cursor, limit)

	if err != nil {
		slog.Error("Get items by collection", "collection_id", collection_id, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.Debug("Items fetched", "collection_id", collection_id, "count", len(items))
	WriteJSON(w, http.StatusOK, items)
}

func (h *ItemHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	item_id, err := ParseID(chi.URLParam(r, "item_id"))
	var req models.UpdateItemReadRequest
	decoding_err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || decoding_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.itemService.UpdateItemRead(item_id, req.IsRead)
	if err != nil {
		slog.Error("Mark item read", "item_id", item_id, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ItemHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateItemsMarkAllRequest
	decoding_err := json.NewDecoder(r.Body).Decode(&req)
	if decoding_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := h.itemService.UpdateItemsRead(req.ItemIDs, req.IsRead)
	if err != nil {
		slog.Error("Mark all items read", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ItemHandler) MarkFavorite(w http.ResponseWriter, r *http.Request) {
	item_id, err := ParseID(chi.URLParam(r, "item_id"))
	var req models.UpdateItemFavoriteRequest
	decoding_err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || decoding_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.itemService.UpdateItemFavorite(item_id, req.IsFavorite)
	if err != nil {
		slog.Error("Mark item favorite", "item_id", item_id, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ItemHandler) Post(w http.ResponseWriter, r *http.Request) {
	// to do later, no usecase for creating individual items
	w.Write([]byte("item post handler"))
}

func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("handler"))
}

func parseItemFilter(r *http.Request) models.ItemFilter {
	filter := models.ItemFilter{}
	if is_read, err := strconv.ParseBool(r.URL.Query().Get("read")); err == nil {
		filter.IsRead = &is_read
	}
	if is_favorite, err := strconv.ParseBool(r.URL.Query().Get("favorite")); err == nil {
		filter.IsFavorite = &is_favorite
	}
	return filter
}
