package repositories

import (
	"database/sql"
	"testing"

	"github.com/Swunci/rss-feed-backend/internal/models"
)

func createTestFeed(t *testing.T, db *sql.DB) models.Feed {
	result, err := db.Exec("INSERT INTO feeds (url, name) VALUES (?, ?)", "https://example.com/feed", "Example")
	if err != nil {
		t.Fatalf("failed to create test feed: %v", err)
	}
	id, _ := result.LastInsertId()
	return models.Feed{ID: int(id), URL: "https://example.com/feed", Name: "Example"}
}

func createTestItem(t *testing.T, db *sql.DB) int {
	result, err := db.Exec(`INSERT INTO feeds (url, name) VALUES ('https://example.com/feed.xml', "Example")`)
	if err != nil {
		t.Fatalf("failed to insert test feed: %v", err)
	}
	id, _ := result.LastInsertId()
	result, err = db.Exec(`INSERT INTO items (feed_id, title, link) VALUES (?, 'Test Item', 'https://example.com/item')`, id)
	if err != nil {
		t.Fatalf("failed to insert test item: %v", err)
	}
	return int(id)
}

func createTestCollection(t *testing.T, db *sql.DB) models.Collection {
	result, err := db.Exec("INSERT INTO collections (name) VALUES (?)", "Test Collection")
	if err != nil {
		t.Fatalf("failed to create test collection: %v", err)
	}
	id, _ := result.LastInsertId()
	return models.Collection{ID: int(id), Name: "Test Collection"}
}

func createTestFeedWithCollection(t *testing.T, db *sql.DB, collectionID int, url string) models.Feed {
	result, err := db.Exec("INSERT INTO feeds (url, name, collection_id) VALUES (?, ?, ?)", url, "Example", collectionID)
	if err != nil {
		t.Fatalf("failed to create test feed: %v", err)
	}
	id, _ := result.LastInsertId()
	return models.Feed{ID: int(id), URL: url, Name: "Example", CollectionID: &collectionID}
}
