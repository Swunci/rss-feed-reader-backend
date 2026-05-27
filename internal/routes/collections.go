package routes

import (
	"github.com/go-chi/chi/v5"
)

func CollectionRoutes(h *Handlers) chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.Collection.Get)
	r.Post("/", h.Collection.Post)
	r.Put("/{collection_id}", h.Collection.Put)
	r.Delete("/{collection_id}", h.Collection.Delete)

	return r
}
