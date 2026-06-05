package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Swunci/rss-feed-backend/internal/models"
	"github.com/Swunci/rss-feed-backend/internal/services"
	"github.com/go-chi/chi/v5"
)

type FeedService interface {
	GetFeed(feed_id int) (models.Feed, error)
	GetAllFeeds() ([]models.Feed, error)
	CreateFeed(url, name string) (models.Feed, error)
	UpdateFeed(feed_id int, url, name *string) error
	DeleteFeed(feed_id int) error
}

type PollingService interface {
	StartFeed(feed models.Feed)
	StopFeed(feed models.Feed)
}

type FeedHandler struct {
	feedService     *services.FeedService
	pollingService  *services.PollingService
	discoverService *services.DiscoverService
}

func NewFeedHandler(fs *services.FeedService, ps *services.PollingService, ds *services.DiscoverService) *FeedHandler {
	return &FeedHandler{feedService: fs, pollingService: ps, discoverService: ds}
}

func (h *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	feed_id, err := ParseID(chi.URLParam(r, "feed_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		println(err.Error())
		return
	}
	feed, err := h.feedService.GetFeed(feed_id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, feed)
}

func (h *FeedHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")

	var feeds []models.FeedResponse
	var err error

	switch filter {
	case string(models.FeedFilterUnread):
		feeds, err = h.feedService.GetAllUnread()
	case string(models.FeedFilterFavorite):
		feeds, err = h.feedService.GetAllFavorite()
	default:
		feeds, err = h.feedService.GetAllFeeds()
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, feeds)
}

func (h *FeedHandler) Post(w http.ResponseWriter, r *http.Request) {
	var req models.FeedPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	feed, err := h.feedService.CreateFeed(req.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.pollingService.StartFeed(feed)
	WriteJSON(w, http.StatusOK, feed)
}

func (h *FeedHandler) RefreshFeeds(w http.ResponseWriter, r *http.Request) {
	err := h.pollingService.RefreshAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FeedHandler) RefreshFeed(w http.ResponseWriter, r *http.Request) {
	feed_id, err := ParseID(chi.URLParam(r, "feed_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.pollingService.RefreshFeed(feed_id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *FeedHandler) Patch(w http.ResponseWriter, r *http.Request) {
	feed_id, err := ParseID(chi.URLParam(r, "feed_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var req models.FeedUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := h.feedService.UpdateFeed(feed_id, req.URL, req.Name, req.CollectionID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		println(err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *FeedHandler) UnassignCollection(w http.ResponseWriter, r *http.Request) {
	feed_id, err := ParseID(chi.URLParam(r, "feed_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := h.feedService.RemoveFeedFromCollection(feed_id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		println(err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *FeedHandler) Delete(w http.ResponseWriter, r *http.Request) {
	feed_id, err := ParseID(chi.URLParam(r, "feed_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		println(err.Error())
		return
	}
	if err := h.feedService.DeleteFeed(feed_id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		println(err.Error())
		return
	}
	h.pollingService.StopFeed(feed_id)
	w.WriteHeader(http.StatusOK)
}

func (h *FeedHandler) Discover(w http.ResponseWriter, r *http.Request) {
	var req models.FeedPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	feed_res, err := h.discoverService.DiscoverFeeds(req.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	WriteJSON(w, http.StatusOK, feed_res)
}
