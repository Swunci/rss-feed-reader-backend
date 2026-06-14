package services

import (
	"log/slog"

	"github.com/Swunci/rss-feed-backend/internal/models"
	"github.com/Swunci/rss-feed-backend/internal/repositories"
	"github.com/mmcdole/gofeed"
)

type FeedRepository interface {
	GetFeed(feed_id int) (models.Feed, error)
	GetAllFeeds() ([]models.Feed, error)
	CreateFeed(url, name string) (models.Feed, error)
	UpdateFeed(feed_id int, url, name *string) error
	DeleteFeed(feed_id int) error
}

type FeedService struct {
	feedRepo *repositories.FeedRepo
	itemRepo *repositories.ItemRepo
}

func NewFeedService(fr *repositories.FeedRepo, ir *repositories.ItemRepo) *FeedService {
	return &FeedService{feedRepo: fr, itemRepo: ir}
}

func (s *FeedService) GetFeed(feed_id int) (models.Feed, error) {
	return s.feedRepo.GetFeed(feed_id)
}
func (s *FeedService) GetAllFeeds() ([]models.FeedResponse, error) {
	return s.feedRepo.GetAllFeedsWithCount()
}

func (s *FeedService) GetAllUnread() ([]models.FeedResponse, error) {
	feed_ids, err := s.itemRepo.GetUnreadItemsFeedIds()
	if err != nil {
		slog.Error("Get unread feed ids", "err", err)
		return []models.FeedResponse{}, err
	}

	feeds, err := s.feedRepo.GetFeeds(feed_ids, models.FeedFilterUnread)
	if err != nil {
		slog.Error("Get unread feeds", "err", err)
		return []models.FeedResponse{}, err
	}
	return feeds, nil
}

func (s *FeedService) GetAllFavorite() ([]models.FeedResponse, error) {
	feed_ids, err := s.itemRepo.GetFavoriteItemsFeedIds()
	if err != nil {
		slog.Error("Get favorite feed ids", "err", err)
		return []models.FeedResponse{}, err
	}

	feeds, err := s.feedRepo.GetFeeds(feed_ids, models.FeedFilterFavorite)
	if err != nil {
		slog.Error("Get favorite feeds", "err", err)
		return []models.FeedResponse{}, err
	}
	return feeds, nil
}

func (s *FeedService) CreateFeed(url, title string) (models.Feed, error) {
	if isRedditURL(url) {
		url = url + "?user=Positive_Ear1287&feed=fa6c8aa5fdc3af2f011b2cdc6cec7be7ec664436"
	}
	fp := gofeed.NewParser()
	parsed, err := fp.ParseURL(url)
	if err != nil {
		slog.Error("Parse feed link", "feed_url", url, "err", err)
		return models.Feed{}, err
	}
	feed_title := parsed.Title
	if len(title) > 0 {
		feed_title = title
	}
	return s.feedRepo.CreateFeed(url, feed_title)
}

func (s *FeedService) UpdateFeed(feed_id int, url, name *string, collection_id *int) error {
	return s.feedRepo.UpdateFeed(feed_id, url, name, collection_id)
}

func (s *FeedService) RemoveFeedFromCollection(feed_id int) error {
	return s.feedRepo.RemoveFeedFromCollection(feed_id)
}

func (s *FeedService) DeleteFeed(feed_id int) error {
	return s.feedRepo.DeleteFeed(feed_id)
}
