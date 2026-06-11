package repositories

import (
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Swunci/rss-feed-backend/internal/database"
	"github.com/Swunci/rss-feed-backend/internal/models"
)

func TestCreateItems(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	err := repo.CreateItems(feed.ID, items)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM items WHERE feed_id = ?", feed.ID).Scan(&count)
	if count != 2 {
		t.Errorf("expected 2 items in db, got %d", count)
	}
}

func TestCreateItems_Empty(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	err := repo.CreateItems(feed.ID, []models.Item{})
	if err != nil {
		t.Fatalf("expected no error for empty items, got %v", err)
	}
}

func TestCreateItems_DuplicateLink(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

	err := repo.CreateItems(feed.ID, items)
	if err != nil {
		t.Fatalf("expected no error for duplicate, got %v", err)
	}

	result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{}, "", 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 item after duplicate insert, got %d", len(result))
	}
}

func TestCreateItems_InvalidFeedID(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewItemRepo(db, db)

	err := repo.CreateItems(99999, []models.Item{
		{FeedID: 99999, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	})
	if err == nil {
		t.Fatal("expected error for invalid feed_id, got nil")
	}
}
