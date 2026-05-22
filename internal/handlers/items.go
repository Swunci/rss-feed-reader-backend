package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Swunci/rrs-feed-backend/internal/models"
	"github.com/Swunci/rrs-feed-backend/internal/services"
	"github.com/go-chi/chi/v5"
)

type ItemHandler struct {
	itemService *services.ItemService
	logger      *slog.Logger
}

func NewItemHandler(s *services.ItemService, logger *slog.Logger) *ItemHandler {
	return &ItemHandler{itemService: s, logger: logger}
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
		w.WriteHeader(http.StatusInternalServerError)
		println(err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *ItemHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	feed_id, err := ParseID(chi.URLParam(r, "feed_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		println(err.Error())
		return
	}
	filter := models.ItemFilter{}
	is_read, err := strconv.ParseBool(r.URL.Query().Get("read"))
	if err == nil {
		filter.IsRead = &is_read
		h.logger.Info("Read filter", "feed_id", feed_id, "value", *filter.IsRead)
	}
	is_favorite, err := strconv.ParseBool(r.URL.Query().Get("favorite"))
	if err == nil {
		filter.IsFavorite = &is_favorite
		h.logger.Info("Favorite filter", "feed_id", feed_id, "value", *filter.IsFavorite)
	}
	items, err := h.itemService.GetItems(feed_id, filter)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		println(err.Error())
		return
	}
	h.logger.Info("Successful items retrieval", "num_of_items", len(items))
	WriteJSON(w, http.StatusOK, items)
}

func (h *ItemHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	item_id, err := ParseID(chi.URLParam(r, "item_id"))
	var req models.UpdateItemReadRequest
	decoding_err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || decoding_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		println(err.Error())
		return
	}
	err = h.itemService.UpdateItemRead(item_id, req.IsRead)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ItemHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateItemsMarkAllRequest
	decoding_err := json.NewDecoder(r.Body).Decode(&req)
	if decoding_err != nil {
		h.logger.Error("Invalid JSON", "err", decoding_err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := h.itemService.UpdateItemsRead(req.ItemIDs, req.IsRead)
	if err != nil {
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
