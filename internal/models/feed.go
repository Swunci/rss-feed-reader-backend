package models

type Feed struct {
	ID           int    `json:"id"`
	URL          string `json:"url"`
	Name         string `json:"name"`
	CollectionID *int   `json:"collection_id"`
}

type FeedPostRequest struct {
	URL  string `json:"url"`
	Name string `json:"name,omitempty"`
}

type FeedUpdateRequest struct {
	URL          *string `json:"url"`
	Name         *string `json:"name"`
	CollectionID *int    `json:"collection_id"`
}

type FeedResponse struct {
	Feed
	Count int `json:"count"`
}

type FeedFilter string

const (
	FeedFilterUnread   FeedFilter = "unread"
	FeedFilterFavorite FeedFilter = "favorite"
)

type DiscoverFeed struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
