package repositories

import (
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Swunci/rss-feed-backend/internal/database"
	"github.com/Swunci/rss-feed-backend/internal/models"
)

func TestGetFeed(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db)

	created, err := repo.CreateFeed("https://example.com/feed", "Example")
	if err != nil {
		t.Fatalf("expected no error creating feed, got %v", err)
	}

	feed, err := repo.GetFeed(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if feed.ID != created.ID {
		t.Errorf("expected id %d, got %d", created.ID, feed.ID)
	}
	if feed.URL != created.URL {
		t.Errorf("expected url %s, got %s", created.URL, feed.URL)
	}
	if feed.Name != created.Name {
		t.Errorf("expected name %s, got %s", created.Name, feed.Name)
	}
}

func TestGetFeed_NotFound(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db)

	_, err := repo.GetFeed(999)
	if err == nil {
		t.Fatal("expected error for non-existent feed, got nil")
	}
}

func TestGetAllFeeds(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db)
	repo.CreateFeed("https://example.com/feed", "Example")
	repo.CreateFeed("https://other.com/feed", "Other")

	feeds, err := repo.GetAllFeeds()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feeds) != 2 {
		t.Errorf("expected 2 feeds, got %d", len(feeds))
	}
}

func TestGetAllFeeds_Empty(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db)

	feeds, err := repo.GetAllFeeds()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feeds) != 0 {
		t.Errorf("expected 0 feeds, got %d", len(feeds))
	}
}

func TestGetAllFeedsWithCount(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feedRepo := NewFeedRepo(db, db)
	itemRepo := NewItemRepo(db, db)

	feed, _ := feedRepo.CreateFeed("https://example.com/feed", "Example")
	itemRepo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/item1", Description: "desc", PublishedAt: time.Now()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/item2", Description: "desc", PublishedAt: time.Now()},
	})

	feeds, err := feedRepo.GetAllFeedsWithCount()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feeds) != 1 {
		t.Errorf("expected 1 feed, got %d", len(feeds))
	}
	if feeds[0].Count != 2 {
		t.Errorf("expected count 2, got %d", feeds[0].Count)
	}
}

func TestGetAllFeedsWithCount_Empty(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db)

	feeds, err := repo.GetAllFeedsWithCount()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feeds) != 0 {
		t.Errorf("expected 0 feeds, got %d", len(feeds))
	}
}

func TestGetAllFeedsWithCount_NoItems(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db)
	repo.CreateFeed("https://example.com/feed", "Example")

	feeds, err := repo.GetAllFeedsWithCount()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if feeds[0].Count != 0 {
		t.Errorf("expected count 0, got %d", feeds[0].Count)
	}
}

func TestGetAllFeedsWithCount_MultipleFeeds(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feedRepo := NewFeedRepo(db, db)
	itemRepo := NewItemRepo(db, db)

	feed1, _ := feedRepo.CreateFeed("https://example.com/feed", "Example")
	feed2, _ := feedRepo.CreateFeed("https://other.com/feed", "Other")

	itemRepo.CreateItems(feed1.ID, []models.Item{
		{FeedID: feed1.ID, Title: "Item 1", Link: "https://example.com/item1", Description: "desc", PublishedAt: time.Now()},
		{FeedID: feed1.ID, Title: "Item 2", Link: "https://example.com/item2", Description: "desc", PublishedAt: time.Now()},
	})
	itemRepo.CreateItems(feed2.ID, []models.Item{
		{FeedID: feed2.ID, Title: "Item 3", Link: "https://other.com/item3", Description: "desc", PublishedAt: time.Now()},
	})

	feeds, err := feedRepo.GetAllFeedsWithCount()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feeds) != 2 {
		t.Fatalf("expected 2 feeds, got %d", len(feeds))
	}

	counts := map[int]int{}
	for _, f := range feeds {
		counts[f.ID] = f.Count
	}
	if counts[feed1.ID] != 2 {
		t.Errorf("expected count 2 for feed1, got %d", counts[feed1.ID])
	}
	if counts[feed2.ID] != 1 {
		t.Errorf("expected count 1 for feed2, got %d", counts[feed2.ID])
	}
}

func TestGetFeeds(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feedRepo := NewFeedRepo(db, db)
	itemRepo := NewItemRepo(db, db)

	feed1, _ := feedRepo.CreateFeed("https://example.com/feed", "Example")
	feed2, _ := feedRepo.CreateFeed("https://other.com/feed", "Other")

	itemRepo.CreateItems(feed1.ID, []models.Item{
		{FeedID: feed1.ID, Title: "Item 1", Link: "https://example.com/item1", Description: "desc", PublishedAt: time.Now()},
	})

	feeds, err := feedRepo.GetFeeds([]int{feed1.ID, feed2.ID}, models.FeedFilterUnread)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feeds) != 2 {
		t.Errorf("expected 2 feeds, got %d", len(feeds))
	}
	if feeds[0].Count != 1 {
		t.Errorf("expected count 1 for feed1, got %d", feeds[0].Count)
	}
	if feeds[1].Count != 0 {
		t.Errorf("expected count 0 for feed2, got %d", feeds[1].Count)
	}
}

func TestGetFeeds_FavoriteFilter(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feedRepo := NewFeedRepo(db, db)
	itemRepo := NewItemRepo(db, db)

	feed, _ := feedRepo.CreateFeed("https://example.com/feed", "Example")
	itemRepo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/item1", Description: "desc", PublishedAt: time.Now()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/item2", Description: "desc", PublishedAt: time.Now()},
	})

	items, _ := itemRepo.GetItemsByFeed(feed.ID, models.ItemFilter{}, "", 0)
	itemRepo.UpdateFavorite(items[0].ID, true)

	feeds, err := feedRepo.GetFeeds([]int{feed.ID}, models.FeedFilterFavorite)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if feeds[0].Count != 1 {
		t.Errorf("expected count 1, got %d", feeds[0].Count)
	}
}

func TestGetFeeds_PartialMatch(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db)

	feed, _ := repo.CreateFeed("https://example.com/feed", "Example")

	feeds, err := repo.GetFeeds([]int{feed.ID, 999}, models.FeedFilterUnread)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feeds) != 1 {
		t.Errorf("expected 1 feed, got %d", len(feeds))
	}
}

func TestGetFeeds_Empty(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db)

	feeds, err := repo.GetFeeds([]int{}, models.FeedFilterUnread)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feeds) != 0 {
		t.Errorf("expected 0 feeds, got %d", len(feeds))
	}
}

func TestGetFeeds_UnreadFilter(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feedRepo := NewFeedRepo(db, db)
	itemRepo := NewItemRepo(db, db)

	feed, _ := feedRepo.CreateFeed("https://example.com/feed", "Example")
	itemRepo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/item1", Description: "desc", PublishedAt: time.Now()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/item2", Description: "desc", PublishedAt: time.Now()},
	})

	items, _ := itemRepo.GetItemsByFeed(feed.ID, models.ItemFilter{}, "", 0)
	itemRepo.UpdateRead(items[0].ID, true)

	feeds, err := feedRepo.GetFeeds([]int{feed.ID}, models.FeedFilterUnread)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if feeds[0].Count != 1 {
		t.Errorf("expected count 1 after marking one read, got %d", feeds[0].Count)
	}
}
