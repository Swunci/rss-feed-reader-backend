package repositories

import (
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Swunci/rss-feed-backend/internal/database"
	"github.com/Swunci/rss-feed-backend/internal/models"
)

func TestCreateFeed(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db, nil)

	feed, err := repo.CreateFeed("https://example.com/feed", "Example")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if feed.URL != "https://example.com/feed" {
		t.Errorf("expected url %s, got %s", "https://example.com/feed", feed.URL)
	}
	if feed.Name != "Example" {
		t.Errorf("expected name %s, got %s", "Example", feed.Name)
	}
}

func TestCreateFeed_DuplicateURL(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db, nil)

	original, err := repo.CreateFeed("https://example.com/feed", "Example")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	duplicate, err := repo.CreateFeed("https://example.com/feed", "Duplicate")
	if err != nil {
		t.Fatalf("expected no error for duplicate url, got %v", err)
	}
	if duplicate.ID != original.ID {
		t.Errorf("expected same feed to be returned, got different IDs %d vs %d", original.ID, duplicate.ID)
	}
}
func TestGetFeed(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db, nil)

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

	repo := NewFeedRepo(db, db, nil)

	_, err := repo.GetFeed(999)
	if err == nil {
		t.Fatal("expected error for non-existent feed, got nil")
	}
}

func TestGetAllFeeds(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db, nil)
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
	repo := NewFeedRepo(db, db, nil)

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

	feedRepo := NewFeedRepo(db, db, nil)
	itemRepo := NewItemRepo(db, db, nil)

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

	repo := NewFeedRepo(db, db, nil)

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

	repo := NewFeedRepo(db, db, nil)
	repo.CreateFeed("https://example.com/feed", "Example")

	feeds, err := repo.GetAllFeedsWithCount()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if feeds[0].Count != 0 {
		t.Errorf("expected count 0, got %d", feeds[0].Count)
	}
}

func TestGetFeeds(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feedRepo := NewFeedRepo(db, db, nil)
	itemRepo := NewItemRepo(db, db, nil)

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

	feedRepo := NewFeedRepo(db, db, nil)
	itemRepo := NewItemRepo(db, db, nil)

	feed, _ := feedRepo.CreateFeed("https://example.com/feed", "Example")
	itemRepo.CreateItems(feed.ID, []models.Item{
		{FeedID: feed.ID, Title: "Item 1", Link: "https://example.com/item1", Description: "desc", PublishedAt: time.Now()},
		{FeedID: feed.ID, Title: "Item 2", Link: "https://example.com/item2", Description: "desc", PublishedAt: time.Now()},
	})

	items, _ := itemRepo.GetItemsByFeed(feed.ID, models.ItemFilter{}, "")
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

	repo := NewFeedRepo(db, db, nil)

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

	repo := NewFeedRepo(db, db, nil)

	feeds, err := repo.GetFeeds([]int{}, models.FeedFilterUnread)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(feeds) != 0 {
		t.Errorf("expected 0 feeds, got %d", len(feeds))
	}
}

func TestUpdateFeed(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()
	repo := NewFeedRepo(db, db, nil)

	newURL := "https://new.com/feed"
	newName := "New Name"

	original_feed, err := repo.CreateFeed("https://example.com/feed", "Example")
	if err != nil {
		t.Fatalf("expected no error creating feed, got %v", err)
	}

	err = repo.UpdateFeed(original_feed.ID, &newURL, &newName)
	if err != nil {
		t.Fatalf("expected no error updating feed, got %v", err)
	}

	return_feed, err := repo.GetFeed(original_feed.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if return_feed.URL != newURL {
		t.Fatalf("URL was not updated correctly, got %v", err)
	}
	if return_feed.Name != newName {
		t.Fatalf("Name was not updated correctly, got %v", err)
	}

}

func TestUpdateFeed_NotFound(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db, nil)

	newName := "New Name"
	err := repo.UpdateFeed(999, nil, &newName)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateFeed_PartialUpdate(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()
	repo := NewFeedRepo(db, db, nil)

	oldUrl := "https://example.com/feed"
	oldName := "Example"
	newName := "New Name"

	feed, err := repo.CreateFeed(oldUrl, oldName)

	if err != nil {
		t.Fatalf("expected no error with createFeed, got %v", err)
	}

	err = repo.UpdateFeed(feed.ID, nil, &newName)
	if err != nil {
		t.Fatalf("expected no error with UpdateFeed, got %v", err)
	}

	updated_feed, err := repo.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("expected no error with GetFeed, got %v", err)
	}
	if updated_feed.URL != oldUrl {
		t.Fatalf("expected url to not change, updatedURL: %v, oldURL: %v", updated_feed.URL, oldUrl)
	}
	if updated_feed.Name == oldName {
		t.Fatalf("expected update to name, updatedName: %v, NewName: %v", updated_feed.Name, newName)
	}
}

func TestUpdateFeed_NilBoth(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db, nil)

	feed, err := repo.CreateFeed("https://example.com/feed", "Example")
	if err != nil {
		t.Fatalf("expected no error creating feed, got %v", err)
	}

	err = repo.UpdateFeed(feed.ID, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	unchanged, err := repo.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if unchanged.URL != feed.URL {
		t.Errorf("expected URL to remain %s, got %s", feed.URL, unchanged.URL)
	}
	if unchanged.Name != feed.Name {
		t.Errorf("expected Name to remain %s, got %s", feed.Name, unchanged.Name)
	}
}

func TestDeleteFeed(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db, nil)

	feed, err := repo.CreateFeed("https://example.com/feed", "Example")
	if err != nil {
		t.Fatalf("expected no error creating feed, got %v", err)
	}

	err = repo.DeleteFeed(feed.ID)
	if err != nil {
		t.Fatalf("expected no error deleting feed, got %v", err)
	}

	_, err = repo.GetFeed(feed.ID)
	if err == nil {
		t.Fatal("expected error for deleted feed, got nil")
	}
}

func TestDeleteFeed_NotFound(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db, nil)

	err := repo.DeleteFeed(999)
	if err != nil {
		t.Fatalf("expected no error deleting non-existent feed, got %v", err)
	}
}

func TestDeleteFeed_CascadesItems(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	feedRepo := NewFeedRepo(db, db, nil)
	itemRepo := NewItemRepo(db, db, nil)

	feed, err := feedRepo.CreateFeed("https://example.com/feed", "Example")
	if err != nil {
		t.Fatalf("expected no error creating feed, got %v", err)
	}

	items := []models.Item{
		{
			FeedID:      feed.ID,
			Title:       "Item 1",
			Link:        "https://example.com/item1",
			Description: "Description 1",
			PublishedAt: time.Now(),
		},
		{
			FeedID:      feed.ID,
			Title:       "Item 2",
			Link:        "https://example.com/item2",
			Description: "Description 2",
			PublishedAt: time.Now(),
		},
	}

	err = itemRepo.CreateItems(feed.ID, items)
	if err != nil {
		t.Fatalf("expected no error creating items, got %v", err)
	}

	err = feedRepo.DeleteFeed(feed.ID)
	if err != nil {
		t.Fatalf("expected no error deleting feed, got %v", err)
	}

	remainingItems, err := itemRepo.GetItemsByFeed(feed.ID, models.ItemFilter{}, "")
	if err != nil {
		t.Fatalf("expected no error fetching items, got %v", err)
	}
	if len(remainingItems) != 0 {
		t.Errorf("expected 0 items after feed deletion, got %d", len(remainingItems))
	}
}
