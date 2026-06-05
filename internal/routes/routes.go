package routes

import (
	"net/http"

	"github.com/Swunci/rss-feed-backend/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handlers struct {
	Item       *handlers.ItemHandler
	ItemSEE    *handlers.ItemSSEHandler
	Feed       *handlers.FeedHandler
	Collection *handlers.CollectionHandler
}

func MainRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Mount("/feeds", FeedRoutes(h))
	r.Mount("/items", ItemRoutes(h))
	r.Mount("/collections", CollectionRoutes(h))

	return r
}
