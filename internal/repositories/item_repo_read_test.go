package repositories

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Swunci/rss-feed-backend/internal/database"
	"github.com/Swunci/rss-feed-backend/internal/models"
)

func TestGetItem(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	items := []models.Item{
		{
			FeedID:      feed.ID,
			Title:       "Full Item",
			Link:        "https://example.com/full",
			Description: "a description",
			PublishedAt: time.Now().UTC(),
		},
	}
	repo.CreateItems(feed.ID, items)

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/full").Scan(&id)
	repo.UpdateRead(id, true)
	repo.UpdateFavorite(id, true)

	item, err := repo.GetItem(id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item.FeedID != feed.ID {
		t.Errorf("expected feed_id %d, got %d", feed.ID, item.FeedID)
	}
	if item.Description != "a description" {
		t.Errorf("expected description 'a description', got %s", item.Description)
	}
	if !item.IsRead {
		t.Error("expected is_read to be true")
	}
	if !item.IsFavorite {
		t.Error("expected is_favorite to be true")
	}
}

func TestGetItem_NotFound(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()
	repo := NewItemRepo(db, db)

	_, err := repo.GetItem(999)
	if err == nil {
		t.Fatal("expected error for non-existent id, got nil")
	}
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetItemsByFeed(t *testing.T) {
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
		t.Fatalf("failed to create items: %v", err)
	}

	result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

func TestGetItemsByFeed_Empty(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	items, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestGetItemsByFeed_FilterIsRead(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Read Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Unread Item", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)
	repo.UpdateRead(id, true)

	isRead := true
	result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{IsRead: &isRead}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 read item, got %d", len(result))
	}
	if result[0].Title != "Read Item" {
		t.Errorf("expected 'Read Item', got %s", result[0].Title)
	}
}

func TestGetItemsByFeed_FilterIsUnread(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Read Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Unread Item", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)
	repo.UpdateRead(id, true)

	isRead := false
	result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{IsRead: &isRead}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 unread item, got %d", len(result))
	}
	if result[0].Title != "Unread Item" {
		t.Errorf("expected 'Unread Item', got %s", result[0].Title)
	}
}

func TestGetItemsByFeed_FilterIsFavorite(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Favorited Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Normal Item", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)
	repo.UpdateFavorite(id, true)

	isFavorite := true
	result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{IsFavorite: &isFavorite}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 favorited item, got %d", len(result))
	}
	if result[0].Title != "Favorited Item" {
		t.Errorf("expected 'Favorited Item', got %s", result[0].Title)
	}
}

func TestGetItemsByFeed_FilterIsNotFavorite(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Favorited Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Normal Item", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)
	repo.UpdateFavorite(id, true)

	isFavorite := false
	result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{IsFavorite: &isFavorite}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 non-favorited item, got %d", len(result))
	}
	if result[0].Title != "Normal Item" {
		t.Errorf("expected 'Normal Item', got %s", result[0].Title)
	}
}

func TestGetItemsByFeed_FilterReadAndFavorite(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Read and Favorited", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Read but not Favorited", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Unread and Favorited", Link: "https://example.com/3", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Unread and not Favorited", Link: "https://example.com/4", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

	var id1, id2, id3 int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id1)
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/2").Scan(&id2)
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/3").Scan(&id3)
	repo.UpdateRead(id1, true)
	repo.UpdateRead(id2, true)
	repo.UpdateFavorite(id1, true)
	repo.UpdateFavorite(id3, true)

	isRead := true
	isFavorite := true
	result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{IsRead: &isRead, IsFavorite: &isFavorite}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 item, got %d", len(result))
	}
	if result[0].Title != "Read and Favorited" {
		t.Errorf("expected 'Read and Favorited', got %s", result[0].Title)
	}
}

func TestGetItemsByFeed_Cursor(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	now := time.Now().UTC()
	items := []models.Item{
		{FeedID: feed.ID, Title: "Oldest", Link: "https://example.com/1", PublishedAt: now.Add(-4 * time.Hour)},
		{FeedID: feed.ID, Title: "Older", Link: "https://example.com/2", PublishedAt: now.Add(-3 * time.Hour)},
		{FeedID: feed.ID, Title: "Middle", Link: "https://example.com/3", PublishedAt: now.Add(-2 * time.Hour)},
		{FeedID: feed.ID, Title: "Newer", Link: "https://example.com/4", PublishedAt: now.Add(-1 * time.Hour)},
		{FeedID: feed.ID, Title: "Newest", Link: "https://example.com/5", PublishedAt: now},
	}
	repo.CreateItems(feed.ID, items)

	t.Run("no cursor returns all items in desc order", func(t *testing.T) {
		result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{}, "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(result) != 5 {
			t.Errorf("expected 5 items, got %d", len(result))
		}
		if result[0].Title != "Newest" {
			t.Errorf("expected first item to be 'Newest', got %s", result[0].Title)
		}
	})

	t.Run("cursor returns only items older than cursor", func(t *testing.T) {
		cursor := now.Add(-1 * time.Hour).Format(time.RFC3339)
		result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{}, cursor)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(result) != 3 {
			t.Errorf("expected 3 items, got %d", len(result))
		}
		if result[0].Title != "Middle" {
			t.Errorf("expected first item to be 'Middle', got %s", result[0].Title)
		}
	})

	t.Run("cursor does not include exact timestamp", func(t *testing.T) {
		cursor := now.Add(-2 * time.Hour).Format(time.RFC3339)
		result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{}, cursor)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		for _, item := range result {
			if item.Title == "Middle" {
				t.Error("cursor item should not be included in results")
			}
		}
	})

	t.Run("cursor at oldest returns empty", func(t *testing.T) {
		cursor := now.Add(-4 * time.Hour).Format(time.RFC3339)
		result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{}, cursor)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected 0 items, got %d", len(result))
		}
	})

	t.Run("cursor combined with filter", func(t *testing.T) {
		var id int
		db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/3").Scan(&id)
		repo.UpdateRead(id, true)

		cursor := now.Format(time.RFC3339)
		isRead := true
		result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{IsRead: &isRead}, cursor)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 read item before cursor, got %d", len(result))
		}
		if result[0].Title != "Middle" {
			t.Errorf("expected 'Middle', got %s", result[0].Title)
		}
	})
}

func TestGetItemsByCollection(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	collection := createTestCollection(t, db)
	feed1 := createTestFeedWithCollection(t, db, collection.ID, "https://example.com/feed1")
	feed2 := createTestFeedWithCollection(t, db, collection.ID, "https://example.com/feed2")
	repo := NewItemRepo(db, db)

	err := repo.CreateItems(feed1.ID, []models.Item{
		{FeedID: feed1.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	})
	if err != nil {
		t.Fatalf("failed to create items: %v", err)
	}
	err = repo.CreateItems(feed2.ID, []models.Item{
		{FeedID: feed2.ID, Title: "Item 2", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	})
	if err != nil {
		t.Fatalf("failed to create items: %v", err)
	}

	result, err := repo.GetItemsByCollection(collection.ID, models.ItemFilter{}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

func TestGetUnreadItemsFeedIds_ReturnsFeedIds(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	err := repo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Unread Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	})
	if err != nil {
		t.Fatalf("failed to create items: %v", err)
	}

	feedIDs, err := repo.GetUnreadItemsFeedIds()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feedIDs) != 1 {
		t.Errorf("expected 1 feed id, got %d", len(feedIDs))
	}
	if feedIDs[0] != feed.ID {
		t.Errorf("expected feed id %d, got %d", feed.ID, feedIDs[0])
	}
}

func TestGetUnreadItemsFeedIds_Empty(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	repo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Read Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	})

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)
	repo.UpdateRead(id, true)

	feedIDs, err := repo.GetUnreadItemsFeedIds()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feedIDs) != 0 {
		t.Errorf("expected 0 feed ids, got %d", len(feedIDs))
	}
}

func TestGetUnreadItemsFeedIds_Distinct(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	repo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	})

	feedIDs, err := repo.GetUnreadItemsFeedIds()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feedIDs) != 1 {
		t.Errorf("expected 1 distinct feed id, got %d", len(feedIDs))
	}
}

func TestGetFavoriteItemsFeedIds_ReturnsFeedIds(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	repo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	})

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)
	repo.UpdateFavorite(id, true)

	feedIDs, err := repo.GetFavoriteItemsFeedIds()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feedIDs) != 1 {
		t.Errorf("expected 1 feed id, got %d", len(feedIDs))
	}
	if feedIDs[0] != feed.ID {
		t.Errorf("expected feed id %d, got %d", feed.ID, feedIDs[0])
	}
}

func TestGetFavoriteItemsFeedIds_Empty(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	repo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	})

	feedIDs, err := repo.GetFavoriteItemsFeedIds()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feedIDs) != 0 {
		t.Errorf("expected 0 feed ids, got %d", len(feedIDs))
	}
}

func TestGetFavoriteItemsFeedIds_Distinct(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	repo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	})

	var id1, id2 int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id1)
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/2").Scan(&id2)
	repo.UpdateFavorite(id1, true)
	repo.UpdateFavorite(id2, true)

	feedIDs, err := repo.GetFavoriteItemsFeedIds()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feedIDs) != 1 {
		t.Errorf("expected 1 distinct feed id, got %d", len(feedIDs))
	}
}

func TestGetItemsByFeed_LimitEnforced(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db)

	items := make([]models.Item, 55)
	for i := range items {
		items[i] = models.Item{
			FeedID:      feed.ID,
			Title:       fmt.Sprintf("Item %d", i),
			Link:        fmt.Sprintf("https://example.com/%d", i),
			PublishedAt: time.Now().UTC().Add(time.Duration(i) * time.Minute),
		}
	}
	repo.CreateItems(feed.ID, items)

	result, err := repo.GetItemsByFeed(feed.ID, models.ItemFilter{}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 50 {
		t.Errorf("expected 50 items (limit), got %d", len(result))
	}
}

func TestGetItemsByCollection_Empty(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	collection := createTestCollection(t, db)
	repo := NewItemRepo(db, db)

	result, err := repo.GetItemsByCollection(collection.ID, models.ItemFilter{}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 items, got %d", len(result))
	}
}

func TestGetItemsByCollection_Filter(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	collection := createTestCollection(t, db)
	feed := createTestFeedWithCollection(t, db, collection.ID, "https://example.com/feed")
	repo := NewItemRepo(db, db)

	repo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Read Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Unread Item", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	})

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)
	repo.UpdateRead(id, true)

	isRead := true
	result, err := repo.GetItemsByCollection(collection.ID, models.ItemFilter{IsRead: &isRead}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 read item, got %d", len(result))
	}
	if result[0].Title != "Read Item" {
		t.Errorf("expected 'Read Item', got %s", result[0].Title)
	}
}

func TestGetItemsByCollection_Cursor(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	collection := createTestCollection(t, db)
	feed := createTestFeedWithCollection(t, db, collection.ID, "https://example.com/feed")
	repo := NewItemRepo(db, db)

	now := time.Now().UTC()
	repo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Oldest", Link: "https://example.com/1", PublishedAt: now.Add(-3 * time.Hour)},
		{FeedID: feed.ID, Title: "Middle", Link: "https://example.com/2", PublishedAt: now.Add(-2 * time.Hour)},
		{FeedID: feed.ID, Title: "Newest", Link: "https://example.com/3", PublishedAt: now.Add(-1 * time.Hour)},
	})

	cursor := now.Add(-1 * time.Hour).Format(time.RFC3339)
	result, err := repo.GetItemsByCollection(collection.ID, models.ItemFilter{}, cursor)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items before cursor, got %d", len(result))
	}
	if result[0].Title != "Middle" {
		t.Errorf("expected first item to be 'Middle', got %s", result[0].Title)
	}
}
