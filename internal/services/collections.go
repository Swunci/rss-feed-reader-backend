package services

import (
	"github.com/Swunci/rss-feed-backend/internal/models"
	"github.com/Swunci/rss-feed-backend/internal/repositories"
)

type CollectionService struct {
	collectionRepo *repositories.CollectionRepo
}

func NewCollectionService(cr *repositories.CollectionRepo) *CollectionService {
	return &CollectionService{collectionRepo: cr}
}

func (s *CollectionService) CreateCollection(name string) (models.Collection, error) {
	return s.collectionRepo.CreateCollection(name)
}

func (s *CollectionService) GetCollections() ([]models.Collection, error) {
	return s.collectionRepo.GetCollections()
}

func (s *CollectionService) UpdateCollection(collection_id int, name string) error {
	return s.collectionRepo.UpdateCollection(collection_id, name)
}

func (s *CollectionService) DeleteCollection(collection_id int) error {
	return s.collectionRepo.DeleteCollection(collection_id)
}
