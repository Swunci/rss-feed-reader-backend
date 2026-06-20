package routes

import (
	"io/fs"
	"net/http"

	rssfeedbackend "github.com/Swunci/rss-feed-backend"

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

func MainRouter(h *Handlers, serveStatic bool) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Use(middleware.Logger)
	r.Route("/api", func(api chi.Router) {
		api.Mount("/feeds", FeedRoutes(h))
		api.Mount("/items", ItemRoutes(h))
		api.Mount("/collections", CollectionRoutes(h))
	})

	if serveStatic {
		publicFS, err := fs.Sub(rssfeedbackend.FrontendAssets, "frontend/dist")
		if err != nil {
			panic(err)
		}
		r.Handle("/*", http.FileServer(http.FS(publicFS)))
	}

	return r
}
