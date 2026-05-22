package models

type Feed struct {
	ID   int    `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
}

type FeedPostRequest struct {
	URL string `json:"url"`
}

type FeedUpdateRequest struct {
	ID   int     `json:"id"`
	URL  *string `json:"url"`
	Name *string `json:"name"`
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
