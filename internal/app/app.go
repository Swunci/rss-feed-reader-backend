package app

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	"github.com/rs/cors"
	_ "modernc.org/sqlite"

	"github.com/Swunci/rrs-feed-backend/internal/database"
	"github.com/Swunci/rrs-feed-backend/internal/handlers"
	"github.com/Swunci/rrs-feed-backend/internal/repositories"
	"github.com/Swunci/rrs-feed-backend/internal/routes"
	"github.com/Swunci/rrs-feed-backend/internal/services"
)

type App struct {
	ReadDB  *sql.DB
	WriteDB *sql.DB
	Router  http.Handler
}

func NewApp() *App {
	readDB, readDB_err := sql.Open("sqlite", "app.db")
	if readDB_err != nil {
		panic(readDB_err)
	}

	writeDB, writeDB_err := sql.Open("sqlite", "app.db")
	if writeDB_err != nil {
		panic(writeDB_err)
	}

	readDB.SetMaxOpenConns(10)
	readDB.SetMaxIdleConns(10)

	writeDB.SetMaxOpenConns(1)
	writeDB.SetMaxIdleConns(1)

	configureSQLite(readDB)
	configureSQLite(writeDB)

	var handler slog.Handler
	if os.Getenv("ENV") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	logger := slog.New(handler)

	feedRepo := repositories.NewFeedRepo(readDB, writeDB, logger)
	itemRepo := repositories.NewItemRepo(readDB, writeDB, logger)
	collectionRepo := repositories.NewCollectionRepo(readDB, writeDB, logger)

	db_table_err := database.Migrate(writeDB)
	if db_table_err != nil {
		panic(db_table_err)
	}

	itemSEEChannel := make(chan string, 10)

	feedService := services.NewFeedService(feedRepo, itemRepo, logger)
	itemService := services.NewItemService(itemRepo)
	collectionService := services.NewCollectionService(collectionRepo)
	pollingService := services.NewPollingService(feedRepo, itemRepo, logger, itemSEEChannel)

	handlers := routes.Handlers{
		Item:       handlers.NewItemHandler(itemService, logger),
		ItemSEE:    handlers.NewItemSSEHandler(itemSEEChannel),
		Feed:       handlers.NewFeedHandler(feedService, pollingService),
		Collection: handlers.NewCollectionHandler(collectionService),
	}

	router := routes.MainRouter(&handlers)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
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
