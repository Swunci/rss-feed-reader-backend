package repositories

import (
	"testing"

	_ "modernc.org/sqlite"

	"github.com/Swunci/rss-feed-backend/internal/database"
)

func TestUpdateFeed(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db)

	newURL := "https://new.com/feed"
	newName := "New Name"

	original_feed, err := repo.CreateFeed("https://example.com/feed", "Example")
	if err != nil {
		t.Fatalf("expected no error creating feed, got %v", err)
	}

	err = repo.UpdateFeed(original_feed.ID, &newURL, &newName, nil)
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

	repo := NewFeedRepo(db, db)

	newName := "New Name"
	err := repo.UpdateFeed(999, nil, &newName, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateFeed_PartialUpdate(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db)

	oldUrl := "https://example.com/feed"
	oldName := "Example"
	newName := "New Name"

	feed, err := repo.CreateFeed(oldUrl, oldName)
	if err != nil {
		t.Fatalf("expected no error with createFeed, got %v", err)
	}

	err = repo.UpdateFeed(feed.ID, nil, &newName, nil)
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

	repo := NewFeedRepo(db, db)

	feed, err := repo.CreateFeed("https://example.com/feed", "Example")
	if err != nil {
		t.Fatalf("expected no error creating feed, got %v", err)
	}

	err = repo.UpdateFeed(feed.ID, nil, nil, nil)
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

func TestUpdateFeed_CollectionID(t *testing.T) {
	db := database.SetupTestDB(t)
	defer db.Close()

	repo := NewFeedRepo(db, db)

	var collectionID int
	err := db.QueryRow("INSERT INTO collections (name) VALUES (?) RETURNING id", "Test Collection").Scan(&collectionID)
	if err != nil {
		t.Fatalf("expected no error creating collection, got %v", err)
	}

	feed, err := repo.CreateFeed("https://example.com/feed", "Example")
	if err != nil {
		t.Fatalf("expected no error creating feed, got %v", err)
	}

	err = repo.UpdateFeed(feed.ID, nil, nil, &collectionID)
	if err != nil {
		t.Fatalf("expected no error updating collection_id, got %v", err)
	}

	updated, err := repo.GetFeed(feed.ID)
	if err != nil {
		t.Fatalf("expected no error getting feed, got %v", err)
	}
	if updated.CollectionID == nil || *updated.CollectionID != collectionID {
		t.Errorf("expected collection_id %d, got %v", collectionID, updated.CollectionID)
	}
}
