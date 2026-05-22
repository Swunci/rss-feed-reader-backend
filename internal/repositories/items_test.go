package repositories

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Swunci/rrs-feed-backend/internal/database"
	"github.com/Swunci/rrs-feed-backend/internal/models"
)

func TestCreateItems(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

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
	repo := NewItemRepo(db, db, nil)

	err := repo.CreateItems(feed.ID, []models.Item{})
	if err != nil {
		t.Fatalf("expected no error for empty items, got %v", err)
	}
}

func TestCreateItems_DuplicateLink(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

	err := repo.CreateItems(feed.ID, items)
	if err != nil {
		t.Fatalf("expected no error for duplicate, got %v", err)
	}

	result, err := repo.GetItems(feed.ID, models.ItemFilter{}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 item after duplicate insert, got %d", len(result))
	}
}

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

func TestGetItem(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{
			FeedID:      feed.ID,
			Title:       "Test Item",
			Link:        "https://example.com/1",
			Description: "desc",
			PublishedAt: time.Now().UTC(),
		},
	}
	err := repo.CreateItems(feed.ID, items)
	if err != nil {
		t.Fatalf("failed to create items: %v", err)
	}

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ? and feed_id = ?", "https://example.com/1", feed.ID).Scan(&id)

	item, err := repo.GetItem(id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item.Title != "Test Item" {
		t.Errorf("expected title %s, got %s", "Test Item", item.Title)
	}
	if item.Link != "https://example.com/1" {
		t.Errorf("expected link %s, got %s", "https://example.com/1", item.Link)
	}
}

func TestGetItem_AllFields(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

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
	repo := NewItemRepo(db, db, nil)

	_, err := repo.GetItem(999)
	if err == nil {
		t.Fatal("expected error for non-existent id, got nil")
	}
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetItems(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	err := repo.CreateItems(feed.ID, items)
	if err != nil {
		t.Fatalf("failed to create items: %v", err)
	}

	result, err := repo.GetItems(feed.ID, models.ItemFilter{}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

func TestGetItems_Empty(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	items, err := repo.GetItems(feed.ID, models.ItemFilter{}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestGetItems_FilterIsRead(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Read Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Unread Item", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)
	repo.UpdateRead(id, true)

	isRead := true
	result, err := repo.GetItems(feed.ID, models.ItemFilter{IsRead: &isRead}, "")
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

func TestGetItems_FilterIsUnread(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Read Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Unread Item", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)
	repo.UpdateRead(id, true)

	isRead := false
	result, err := repo.GetItems(feed.ID, models.ItemFilter{IsRead: &isRead}, "")
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

func TestGetItems_FilterIsFavorite(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Favorited Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Normal Item", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)
	repo.UpdateFavorite(id, true)

	isFavorite := true
	result, err := repo.GetItems(feed.ID, models.ItemFilter{IsFavorite: &isFavorite}, "")
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

func TestGetItems_FilterIsNotFavorite(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Favorited Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Normal Item", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)
	repo.UpdateFavorite(id, true)

	isFavorite := false
	result, err := repo.GetItems(feed.ID, models.ItemFilter{IsFavorite: &isFavorite}, "")
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

func TestGetItems_FilterReadAndFavorite(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

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
	result, err := repo.GetItems(feed.ID, models.ItemFilter{IsRead: &isRead, IsFavorite: &isFavorite}, "")
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

func TestGetItems_Cursor(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

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
		result, err := repo.GetItems(feed.ID, models.ItemFilter{}, "")
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
		result, err := repo.GetItems(feed.ID, models.ItemFilter{}, cursor)
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
		result, err := repo.GetItems(feed.ID, models.ItemFilter{}, cursor)
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
		result, err := repo.GetItems(feed.ID, models.ItemFilter{}, cursor)
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
		result, err := repo.GetItems(feed.ID, models.ItemFilter{IsRead: &isRead}, cursor)
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

func TestGetUnreadItemsFeedIds_ReturnsFeedIds(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Unread Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	}
	err := repo.CreateItems(feed.ID, items)
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
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Read Item", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

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
	repo := NewItemRepo(db, db, nil)

	// Two unread items from the same feed
	items := []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

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
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

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
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

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
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

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

	items := []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

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

	items := []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/2", PublishedAt: time.Now().UTC()},
	}
	repo.CreateItems(feed.ID, items)

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

func TestDeleteItem(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feed := createTestFeed(t, db)
	repo := NewItemRepo(db, db, nil)

	items := []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/1", PublishedAt: time.Now().UTC()},
	}
	err := repo.CreateItems(feed.ID, items)
	if err != nil {
		t.Fatalf("failed to create items: %v", err)
	}

	var id int
	db.QueryRow("SELECT id FROM items WHERE link = ?", "https://example.com/1").Scan(&id)

	err = repo.DeleteItem(id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result, err := repo.GetItems(feed.ID, models.ItemFilter{}, "")
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
