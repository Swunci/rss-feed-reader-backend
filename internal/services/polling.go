package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Swunci/rrs-feed-backend/internal/models"

	"github.com/mmcdole/gofeed"
)

type PollingFeedRepository interface {
	GetAllFeeds() ([]models.Feed, error)
	GetFeed(feed_id int) (models.Feed, error)
}

type PollingItemRepository interface {
	CreateItems(feedID int, items []models.Item) error
}

type PollingService struct {
	feedRepo       PollingFeedRepository
	itemRepo       PollingItemRepository
	cancelFuncs    map[int]context.CancelFunc
	logger         *slog.Logger
	itemSEEChannel chan string
}

func NewPollingService(feedRepo PollingFeedRepository, itemRepo PollingItemRepository, logger *slog.Logger, itemSEEChannel chan string) *PollingService {
	if logger == nil {
		logger = slog.Default()
	}
	return &PollingService{
		feedRepo:       feedRepo,
		itemRepo:       itemRepo,
		cancelFuncs:    make(map[int]context.CancelFunc),
		logger:         logger,
		itemSEEChannel: itemSEEChannel,
	}
}

func (s *PollingService) Start() {
	feeds, err := s.feedRepo.GetAllFeeds()
	if err != nil {
		s.logger.Error("Failed to get feeds", "err", err)
		return
	}
	for _, feed := range feeds {
		s.StartFeed(feed)
	}
}

func (s *PollingService) StartFeed(feed models.Feed) {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFuncs[feed.ID] = cancel
	go s.pollFeed(feed, ctx)
}

func (s *PollingService) StopFeed(feed_id int) {
	if cancel, ok := s.cancelFuncs[feed_id]; ok {
		cancel()
		delete(s.cancelFuncs, feed_id)
		s.logger.Info("Stopping feed", "feed_id", feed_id)
	}
}

func (s *PollingService) pollFeed(feed models.Feed, ctx context.Context) {
	interval := time.Duration(15) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.logger.Info("Polling start", "feed_url", feed.URL, "feed_id", feed.ID)
	s.fetchItems(feed)
	for {
		select {
		case <-ticker.C:
			s.fetchItems(feed)
		case <-ctx.Done():
			return
		}
	}
}

func (s *PollingService) fetchItems(feed models.Feed) {
	fp := gofeed.NewParser()
	parsed, err := fp.ParseURL(feed.URL)
	if err != nil {
		s.logger.Error("Parse feed", "feed_url", feed.URL, "err", err)
		return
	}
	s.logger.Info("Items fetched", "feed_url", feed.URL)
	var items []models.Item
	for _, entry := range parsed.Items {
		description := entry.Content
		title := entry.Title
		if description == "" {
			description = entry.Description
		}
		if description == "" {
			description = getYouTubeDescription(entry)
		}
		if title == "" {
			title = "Untitled"
		}

		published_at := time.Unix(0, 0).UTC()

		if entry.PublishedParsed != nil {
			published_at = entry.PublishedParsed.UTC()
		}
		items = append(items, models.Item{
			FeedID:      feed.ID,
			Title:       title,
			Link:        entry.Link,
			Description: description,
			PublishedAt: published_at,
		})
	}
	s.logger.Debug("Items for DB", slog.Any("items", items))

	if err := s.itemRepo.CreateItems(feed.ID, items); err != nil {
		s.logger.Error("Create items", "feed_url", feed.URL, "err", err)
		return
	}
	msg := fmt.Sprintf(`{"feedId": %d}`, feed.ID)
	s.logger.Info("Send item server event", "msg", msg)
	s.itemSEEChannel <- msg
}

func (s *PollingService) RefreshAll() error {
	feeds, err := s.feedRepo.GetAllFeeds()
	if err != nil {
		s.logger.Error("Refresh all feeds", "err", err)
		return err
	}
	for _, feed := range feeds {
		go s.fetchItems(feed)
	}
	return nil
}

func (s *PollingService) RefreshFeed(feed_id int) error {
	feed, err := s.feedRepo.GetFeed(feed_id)
	if err != nil {
		s.logger.Error("Refresh feed", "feed_id", feed_id, "err", err)
		return err
	}
	go s.fetchItems(feed)
	return nil
}

func getYouTubeDescription(entry *gofeed.Item) string {
	if media, ok := entry.Extensions["media"]; ok {
		if group, ok := media["group"]; ok {
			if len(group) > 0 {
				if desc, ok := group[0].Children["description"]; ok {
					if len(desc) > 0 {
						return desc[0].Value
					}
				}
			}
		}
	}
	return ""
}
