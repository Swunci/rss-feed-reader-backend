package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/Swunci/rss-feed-backend/internal/models"

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
	itemSEEChannel chan string
}

func NewPollingService(feedRepo PollingFeedRepository, itemRepo PollingItemRepository, itemSEEChannel chan string) *PollingService {
	return &PollingService{
		feedRepo:       feedRepo,
		itemRepo:       itemRepo,
		cancelFuncs:    make(map[int]context.CancelFunc),
		itemSEEChannel: itemSEEChannel,
	}
}

func (s *PollingService) Start() {
	feeds, err := s.feedRepo.GetAllFeeds()
	if err != nil {
		slog.Error("Get feeds on start", "err", err)
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
		slog.Info("Stopping feed", "feed_id", feed_id)
		cancel()
		delete(s.cancelFuncs, feed_id)

	}
}

func (s *PollingService) pollFeed(feed models.Feed, ctx context.Context) {
	minutes, err := strconv.Atoi(os.Getenv("POLLING_INTERVAL_MINUTES"))
	if err != nil {
		minutes = 15
	}
	interval := time.Duration(minutes) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	slog.Info("Polling start", "feed_url", feed.URL, "feed_id", feed.ID)
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
		slog.Error("Parse feed", "feed_url", feed.URL, "err", err)
		return
	}
	slog.Info("Items fetched", "feed_url", feed.URL)
	var items []models.Item
	for _, entry := range parsed.Items {
		description := entry.Content
		title := entry.Title
		if description == "" {
			description = entry.Description
		}
		if description == "" {
			description = getYouTubeDescription(entry)
			thumbnail := getYouTubeThumbnail(entry)
			if thumbnail != "" {
				description = fmt.Sprintf(`<img src="%s"><br>%s`, thumbnail, description)
			}
		}
		if title == "" {
			title = "Untitled"
		}

		published_at := time.Unix(0, 0).UTC()

		if entry.PublishedParsed != nil {
			published_at = entry.PublishedParsed.UTC()
		}
		item := models.Item{
			FeedID:      feed.ID,
			Title:       title,
			Link:        entry.Link,
			Description: description,
			PublishedAt: published_at,
		}
		items = append(items, item)
		slog.Debug("Item for DB",
			slog.String("title", item.Title),
			slog.String("link", item.Link),
			slog.Time("published_at", item.PublishedAt),
		)

	}

	if err := s.itemRepo.CreateItems(feed.ID, items); err != nil {
		slog.Error("Create items", "feed_url", feed.URL, "err", err)
		return
	}
	msg := fmt.Sprintf(`{"feedId": %d}`, feed.ID)
	s.itemSEEChannel <- msg
}

func (s *PollingService) RefreshAll() error {
	feeds, err := s.feedRepo.GetAllFeeds()
	if err != nil {
		slog.Error("Refresh all feeds", "err", err)
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
		slog.Error("Refresh feed", "feed_id", feed_id, "err", err)
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

func getYouTubeThumbnail(entry *gofeed.Item) string {
	if media, ok := entry.Extensions["media"]; ok {
		if group, ok := media["group"]; ok {
			if len(group) > 0 {
				if thumbnail, ok := group[0].Children["thumbnail"]; ok {
					if len(thumbnail) > 0 {
						return thumbnail[0].Attrs["url"]
					}
				}
			}
		}
	}
	return ""
}
