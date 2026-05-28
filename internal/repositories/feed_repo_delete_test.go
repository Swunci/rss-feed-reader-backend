package repositories

import (
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Swunci/rss-feed-backend/internal/database"
	"github.com/Swunci/rss-feed-backend/internal/models"
)

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
