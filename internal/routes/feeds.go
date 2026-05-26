package routes

import (
	"github.com/go-chi/chi/v5"
)

func FeedRoutes(h *Handlers) chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.Feed.GetAll)
	r.Get("/{feed_id}/items", h.Item.GetItems)
	r.Get("/{feed_id}", h.Feed.GetFeed)
	r.Post("/", h.Feed.Post)
	r.Post("/refresh", h.Feed.RefreshFeeds)
	r.Post("/{feed_id}/refresh", h.Feed.RefreshFeed)
	r.Patch("/{feed_id}", h.Feed.Patch)
	r.Delete("/{feed_id}", h.Feed.Delete)

	return r
}
