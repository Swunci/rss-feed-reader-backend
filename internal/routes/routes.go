package routes

import (
	"net/http"

	"github.com/Swunci/rrs-feed-backend/internal/handlers"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Item    *handlers.ItemHandler
	ItemSEE *handlers.ItemSSEHandler
	Feed    *handlers.FeedHandler
}

func MainRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()

	r.Mount("/feeds", FeedRoutes(h))
	r.Mount("/items", ItemRoutes(h))

	return r
}
