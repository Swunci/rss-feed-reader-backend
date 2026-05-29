package services

import (
	"github.com/Swunci/rss-feed-backend/internal/models"
	"github.com/Swunci/rss-feed-backend/internal/repositories"
)

type ItemRepository interface {
	GetItem(id int) (models.Item, error)
	GetItems(feedID int) ([]models.Item, error)
	CreateItems(feedID int, items []models.Item) error
	DeleteItem(id int) error
}

type ItemService struct {
	repository *repositories.ItemRepo
}

func NewItemService(r *repositories.ItemRepo) *ItemService {
	return &ItemService{repository: r}
}

func (s *ItemService) GetItem(item_id int) (models.Item, error) {
	return s.repository.GetItem(item_id)
}

func (s *ItemService) GetAllItems(filter models.ItemFilter, timestamp_cursor string) ([]models.Item, error) {
	return s.repository.GetAllItems(filter, timestamp_cursor)
}

func (s *ItemService) GetItemsByFeed(feed_id int, filter models.ItemFilter, timestamp_cursor string) ([]models.Item, error) {
	return s.repository.GetItemsByFeed(feed_id, filter, timestamp_cursor)
}

func (s *ItemService) GetItemsByCollection(collection_id int, filter models.ItemFilter, timestamp_cursor string) ([]models.Item, error) {
	return s.repository.GetItemsByCollection(collection_id, filter, timestamp_cursor)
}

func (s *ItemService) CreateItems(feed_id int, items []models.Item) error {
	return s.repository.CreateItems(feed_id, items)
}

func (s *ItemService) UpdateItemRead(item_id int, is_read bool) error {
	return s.repository.UpdateRead(item_id, is_read)
}

func (s *ItemService) UpdateItemsRead(item_ids []int, is_read bool) error {
	return s.repository.UpdateReadMultiple(item_ids, is_read)
}

func (s *ItemService) UpdateItemFavorite(item_id int, is_favorite bool) error {
	return s.repository.UpdateFavorite(item_id, is_favorite)
}

func (s *ItemService) DeleteItem(item_id int) error {
	return s.repository.DeleteItem(item_id)
}
