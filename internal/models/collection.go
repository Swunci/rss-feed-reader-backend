package models

type Collection struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CollectionRequest struct {
	Name string `json:"name"`
}
