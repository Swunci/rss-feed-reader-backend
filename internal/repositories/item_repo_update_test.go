package repositories

import (
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Swunci/rss-feed-backend/internal/database"
	"github.com/Swunci/rss-feed-backend/internal/models"
)

func TestUpdateRead(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewItemRepo(db, db, nil)
	itemID := createTestItem(t, db)

	t.Run("mark as read", func(t *testing.T) {
		err := repo.UpdateRead(itemID, true)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		var isRead bool
		db.QueryRow("SELECT is_read FROM items WHERE id = ?", itemID).Scan(&isRead)
		if !isRead {
			t.Error("expected is_read to be true")
		}
	})

	t.Run("mark as unread", func(t *testing.T) {
		err := repo.UpdateRead(itemID, false)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		var isRead bool
		db.QueryRow("SELECT is_read FROM items WHERE id = ?", itemID).Scan(&isRead)
		if isRead {
			t.Error("expected is_read to be false")
		}
	})

	t.Run("non existent item", func(t *testing.T) {
		err := repo.UpdateRead(99999, true)
		if err != nil {
			t.Fatalf("expected no error for non existent item, got: %v", err)
		}
	})
}

func TestUpdateFavorite(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewItemRepo(db, db, nil)
	itemID := createTestItem(t, db)

	t.Run("mark as favorite", func(t *testing.T) {
		err := repo.UpdateFavorite(itemID, true)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		var isFavorite bool
		db.QueryRow("SELECT is_favorite FROM items WHERE id = ?", itemID).Scan(&isFavorite)
		if !isFavorite {
			t.Error("expected is_favorite to be true")
		}
	})

	t.Run("mark as unfavorite", func(t *testing.T) {
		err := repo.UpdateFavorite(itemID, false)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		var isFavorite bool
		db.QueryRow("SELECT is_favorite FROM items WHERE id = ?", itemID).Scan(&isFavorite)
		if isFavorite {
			t.Error("expected is_favorite to be false")
		}
	})

	t.Run("non existent item", func(t *testing.T) {
		err := repo.UpdateFavorite(99999, true)
		if err != nil {
			t.Fatalf("expected no error for non existent item, got: %v", err)
		}
	})
}

func TestUpdateReadMultiple_MarkAsRead(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	repo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	})

	var id1, id2 int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id1)
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/2").Scan(&id2)

	err := repo.UpdateReadMultiple([]int{id1, id2}, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var isRead1, isRead2 bool
	db.QueryRow("SELECT is_read FROM items WHERE id = ?", id1).Scan(&isRead1)
	db.QueryRow("SELECT is_read FROM items WHERE id = ?", id2).Scan(&isRead2)
	if !isRead1 || !isRead2 {
		t.Error("expected both items to be marked as read")
	}
}

func TestUpdateReadMultiple_MarkAsUnread(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	repo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	})

	var id1, id2 int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id1)
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/2").Scan(&id2)

	repo.UpdateReadMultiple([]int{id1, id2}, true)

	err := repo.UpdateReadMultiple([]int{id1, id2}, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var isRead1, isRead2 bool
	db.QueryRow("SELECT is_read FROM items WHERE id = ?", id1).Scan(&isRead1)
	db.QueryRow("SELECT is_read FROM items WHERE id = ?", id2).Scan(&isRead2)
	if isRead1 || isRead2 {
		t.Error("expected both items to be marked as unread")
	}
}

func TestUpdateReadMultiple_NonExistentIDs(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewItemRepo(db, db, nil)

	err := repo.UpdateReadMultiple([]int{99999, 88888}, true)
	if err != nil {
		t.Fatalf("expected no error for non-existent ids, got %v", err)
	}
}

func TestUpdateReadMultiple_Empty(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewItemRepo(db, db, nil)

	err := repo.UpdateReadMultiple([]int{}, true)
	if err != nil {
		t.Fatalf("expected no error for empty slice, got %v", err)
	}
}
