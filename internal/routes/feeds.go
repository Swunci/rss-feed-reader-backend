package routes

import (
	"time"

	"github.com/Swunci/rss-feed-backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	"golang.org/x/time/rate"
)

func FeedRoutes(h *Handlers) chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.Feed.GetAll)
	r.Get("/{feed_id}/items", h.Item.GetItemsByFeed)
	r.Get("/{feed_id}", h.Feed.GetFeed)
	r.Post("/", h.Feed.Post)
	r.Post("/discover", h.Feed.Discover)
	r.With(middleware.RateLimit(rate.NewLimiter(rate.Every(5*time.Minute), 1))).Post("/refresh", h.Feed.RefreshFeeds)
	r.Post("/{feed_id}/refresh", h.Feed.RefreshFeed)
	r.Patch("/{feed_id}", h.Feed.Patch)
	r.Delete("/{feed_id}", h.Feed.Delete)
	r.Delete("/{feed_id}/unassign", h.Feed.UnassignCollection)

	return r
}
