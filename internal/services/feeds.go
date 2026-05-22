package services

import (
	"log/slog"

	"github.com/Swunci/rrs-feed-backend/internal/models"
	"github.com/Swunci/rrs-feed-backend/internal/repositories"
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
	logger   *slog.Logger
}

func NewFeedService(fr *repositories.FeedRepo, ir *repositories.ItemRepo, logger *slog.Logger) *FeedService {
	return &FeedService{feedRepo: fr, itemRepo: ir, logger: logger}
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
		s.logger.Error("Get unread items' feed_ids", "err", err)
		return []models.FeedResponse{}, err
	}

	feeds, err := s.feedRepo.GetFeeds(feed_ids, models.FeedFilterUnread)
	if err != nil {
		s.logger.Error("Get feeds", "err", err)
		return []models.FeedResponse{}, err
	}
	return feeds, nil
}

func (s *FeedService) GetAllFavorite() ([]models.FeedResponse, error) {
	feed_ids, err := s.itemRepo.GetFavoriteItemsFeedIds()
	if err != nil {
		s.logger.Error("Get favorite items' feed_ids", "err", err)
		return []models.FeedResponse{}, err
	}

	feeds, err := s.feedRepo.GetFeeds(feed_ids, models.FeedFilterFavorite)
	if err != nil {
		s.logger.Error("Get feeds", "err", err)
		return []models.FeedResponse{}, err
	}
	return feeds, nil
}

func (s *FeedService) CreateFeed(url string) (models.Feed, error) {
	fp := gofeed.NewParser()
	parsed, err := fp.ParseURL(url)
	if err != nil {
		s.logger.Error("Parse feed link", "feed_url", url, "err", err)
		return models.Feed{}, err
	}
	return s.feedRepo.CreateFeed(url, parsed.Title)
}

func (s *FeedService) UpdateFeed(feed_id int, url, name *string) error {
	return s.feedRepo.UpdateFeed(feed_id, url, name)
}

func (s *FeedService) DeleteFeed(feed_id int) error {
	return s.feedRepo.DeleteFeed(feed_id)
}
