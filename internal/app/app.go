package app

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
	_ "modernc.org/sqlite"

	"github.com/Swunci/rss-feed-backend/internal/database"
	"github.com/Swunci/rss-feed-backend/internal/handlers"
	"github.com/Swunci/rss-feed-backend/internal/repositories"
	"github.com/Swunci/rss-feed-backend/internal/routes"
	"github.com/Swunci/rss-feed-backend/internal/services"
)

type App struct {
	ReadDB  *sql.DB
	WriteDB *sql.DB
	Router  http.Handler
}

func NewApp(serveStatic bool) *App {
	var handler slog.Handler
	if os.Getenv("APP_ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	readDB, readDB_err := sql.Open("sqlite", getDBPath())
	if readDB_err != nil {
		panic(readDB_err)
	}

	writeDB, writeDB_err := sql.Open("sqlite", getDBPath())
	if writeDB_err != nil {
		panic(writeDB_err)
	}

	readDB.SetMaxOpenConns(10)
	readDB.SetMaxIdleConns(10)

	writeDB.SetMaxOpenConns(1)
	writeDB.SetMaxIdleConns(1)

	configureSQLite(readDB)
	configureSQLite(writeDB)

	feedRepo := repositories.NewFeedRepo(readDB, writeDB)
	itemRepo := repositories.NewItemRepo(readDB, writeDB)
	collectionRepo := repositories.NewCollectionRepo(readDB, writeDB)

	db_table_err := database.Migrate(writeDB)
	if db_table_err != nil {
		panic(db_table_err)
	}

	itemSEEChannel := make(chan string, 10)

	feedService := services.NewFeedService(feedRepo, itemRepo)
	itemService := services.NewItemService(itemRepo)
	discoverService := services.NewDiscoverService(feedRepo, itemRepo, feedService)
	collectionService := services.NewCollectionService(collectionRepo)
	pollingService := services.NewPollingService(feedRepo, itemRepo, itemSEEChannel)

	handlers := routes.Handlers{
		Item:       handlers.NewItemHandler(itemService),
		ItemSEE:    handlers.NewItemSSEHandler(itemSEEChannel),
		Feed:       handlers.NewFeedHandler(feedService, pollingService, discoverService),
		Collection: handlers.NewCollectionHandler(collectionService),
	}

	router := routes.MainRouter(&handlers, serveStatic)
	origins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	pollingService.Start()

	return &App{
		ReadDB:  readDB,
		WriteDB: writeDB,
		Router:  c.Handler(router),
	}
}

func configureSQLite(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA foreign_keys = ON",
	}

	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return err
		}
	}
	return nil
}

func getDBPath() string {
	if path := os.Getenv("DB_PATH"); path != "" {
		return path
	}
	return "./app.db"
}
