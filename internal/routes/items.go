package routes

import (
	"github.com/go-chi/chi/v5"
)

func ItemRoutes(h *Handlers) chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.Item.GetAllItems)
	r.Get("/{item_id}", h.Item.Get)
	r.Get("/events", h.ItemSEE.ItemEvents)
	r.Post("/", h.Item.Post)
	r.Patch("/{item_id}/read", h.Item.MarkRead)
	r.Patch("/read", h.Item.MarkAllRead)
	r.Patch("/{item_id}/favorite", h.Item.MarkFavorite)
	r.Delete("/{item_id}", h.Item.Delete)

	return r
}
