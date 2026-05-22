package models

import "time"

type Item struct {
	ID          int       `json:"id"`
	FeedID      int       `json:"feed_id"`
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description"`
	PublishedAt time.Time `json:"published_at"`
	IsRead      bool      `json:"is_read"`
	IsFavorite  bool      `json:"is_favorite"`
}

type CreateItemRequest struct {
	FeedID      int       `json:"feed_id"`
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description"`
	PublishedAt time.Time `json:"published_at"`
}

type UpdateItemReadRequest struct {
	IsRead bool `json:"is_read"`
}

type UpdateItemFavoriteRequest struct {
	IsFavorite bool `json:"is_favorite"`
}

type UpdateItemsMarkAllRequest struct {
	ItemIDs []int `json:"item_ids"`
	IsRead  bool  `json:"is_read"`
}

type ItemFilter struct {
	IsRead     *bool
	IsFavorite *bool
}
