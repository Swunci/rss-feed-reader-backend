package repositories

import (
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Swunci/rss-feed-backend/internal/database"
	"github.com/Swunci/rss-feed-backend/internal/models"
)

func TestDeleteItem(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	err := repo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	})
	if err != nil {
		t.Fatalf("failed to create items: %v", err)
	}

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)

	err = repo.DeleteItem(id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(result))
	}
}

func TestDeleteItem_NonExistent(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewItemRepo(db, db, nil)

	err := repo.DeleteItem(999)
	if err != nil {
		t.Fatalf("expected no error for non-existent id, got %v", err)
	}
}
