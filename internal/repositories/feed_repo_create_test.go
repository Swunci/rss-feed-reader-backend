package repositories

import (
	"testing"

	_ "modernc.org/sqlite"

	"github.com/Swunci/rss-feed-backend/internal/database"
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

func TestCreateFeed_EmptyFields(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db, nil)

	_, err := repo.CreateFeed("", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
