package services

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/Swunci/rss-feed-backend/internal/models"
)

type mockFeedRepo struct {
	feeds []models.Feed
	feed  models.Feed
	err   error
}

func (m *mockFeedRepo) GetAllFeeds() ([]models.Feed, error) {
	return m.feeds, m.err
}

func (m *mockFeedRepo) GetFeed(feed_id int) (models.Feed, error) {
	return m.feed, m.err
}

type mockItemRepo struct {
	createdItems []models.Item
	err          error
}

func (m *mockItemRepo) CreateItems(feedID int, items []models.Item) error {
	m.createdItems = append(m.createdItems, items...)
	return m.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestStart_StartsPollingForAllFeeds(t *testing.T) {
	feedRepo := &mockFeedRepo{
		feeds: []models.Feed{
			{ID: 1, Name: "Feed 1", URL: "https://example.com/feed1"},
			{ID: 2, Name: "Feed 2", URL: "https://example.com/feed2"},
		},
	}
	itemRepo := &mockItemRepo{}
	svc := NewPollingService(feedRepo, itemRepo, testLogger(), make(chan string))

	svc.Start()
	time.Sleep(100 * time.Millisecond)

	if len(svc.cancelFuncs) != 2 {
		t.Errorf("expected 2 cancel funcs, got %d", len(svc.cancelFuncs))
	}

	svc.StopFeed(1)
	svc.StopFeed(2)
}

func TestStart_ErrorGettingFeeds(t *testing.T) {
	feedRepo := &mockFeedRepo{err: errors.New("db error")}
	itemRepo := &mockItemRepo{}
	svc := NewPollingService(feedRepo, itemRepo, testLogger(), make(chan string))

	svc.Start()

	if len(svc.cancelFuncs) != 0 {
		t.Errorf("expected 0 cancel funcs, got %d", len(svc.cancelFuncs))
	}
}

func TestStartFeed_StoresCancelFunc(t *testing.T) {
	feedRepo := &mockFeedRepo{}
	itemRepo := &mockItemRepo{}
	svc := NewPollingService(feedRepo, itemRepo, testLogger(), make(chan string))

	feed := models.Feed{ID: 1, Name: "Feed 1", URL: "https://example.com/feed"}
	svc.StartFeed(feed)
	time.Sleep(100 * time.Millisecond)

	if _, ok := svc.cancelFuncs[1]; !ok {
		t.Error("expected cancel func for feed 1")
	}

	svc.StopFeed(1)
}

func TestStopFeed_RemovesCancelFunc(t *testing.T) {
	feedRepo := &mockFeedRepo{}
	itemRepo := &mockItemRepo{}
	svc := NewPollingService(feedRepo, itemRepo, testLogger(), make(chan string))

	feed := models.Feed{ID: 1, Name: "Feed 1", URL: "https://example.com/feed"}
	svc.StartFeed(feed)
	time.Sleep(100 * time.Millisecond)

	svc.StopFeed(1)
	time.Sleep(100 * time.Millisecond)

	if _, ok := svc.cancelFuncs[1]; ok {
		t.Error("expected cancel func to be removed")
	}
}

func TestStopFeed_NonExistent(t *testing.T) {
	feedRepo := &mockFeedRepo{}
	itemRepo := &mockItemRepo{}
	svc := NewPollingService(feedRepo, itemRepo, testLogger(), make(chan string))

	svc.StopFeed(999) // should not panic
}
