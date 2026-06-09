package repositories

import (
	"database/sql"
	"fmt"

	"github.com/Swunci/rss-feed-backend/internal/models"
)

type FeedRepo struct {
	readDB  *sql.DB
	writeDB *sql.DB
}

func NewFeedRepo(readDB *sql.DB, writeDB *sql.DB) *FeedRepo {
	return &FeedRepo{readDB: readDB, writeDB: writeDB}
}

func (r *FeedRepo) CreateFeed(url, name string) (models.Feed, error) {
	_, err := r.writeDB.Exec(
		"INSERT OR IGNORE INTO feeds (url, name) VALUES (?, ?)",
		url, name,
	)
	if err != nil {
		return models.Feed{}, err
	}
	var feed = models.Feed{}
	err = r.writeDB.QueryRow("SELECT * FROM feeds WHERE url = ?", url).
		Scan(&feed.ID, &feed.URL, &feed.Name, &feed.CollectionID)

	return feed, err
}

func (r *FeedRepo) GetFeed(feed_id int) (models.Feed, error) {
	row := r.readDB.QueryRow(
		`SELECT * FROM feeds where id = ?`, feed_id,
	)
	var feed models.Feed

	err := row.Scan(
		&feed.ID,
		&feed.URL,
		&feed.Name,
		&feed.CollectionID,
	)
	if err != nil {
		return models.Feed{}, err
	}
	return feed, nil
}

func (r *FeedRepo) GetAllFeeds() ([]models.Feed, error) {
	rows, err := r.readDB.Query("SELECT * FROM feeds")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds = []models.Feed{}
	for rows.Next() {
		var feed models.Feed
		err := rows.Scan(&feed.ID, &feed.URL, &feed.Name, &feed.CollectionID)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}
	return feeds, nil
}

func (r *FeedRepo) GetAllFeedsWithCount() ([]models.FeedResponse, error) {
	rows, err := r.readDB.Query(`
		SELECT feeds.id, feeds.url, feeds.name, feeds.collection_id, COUNT(items.id) as count
		FROM feeds
		LEFT JOIN items ON items.feed_id = feeds.id
		GROUP BY feeds.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds = []models.FeedResponse{}
	for rows.Next() {
		var feed models.FeedResponse
		err := rows.Scan(&feed.ID, &feed.URL, &feed.Name, &feed.CollectionID, &feed.Count)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}
	return feeds, nil
}

func (r *FeedRepo) GetFeeds(feed_ids []int, filter models.FeedFilter) ([]models.FeedResponse, error) {
	count_condition := "items.is_read = false"
	if filter == models.FeedFilterFavorite {
		count_condition = "items.is_favorite = true"
	}

	query := fmt.Sprintf(`
		SELECT feeds.id, feeds.url, feeds.name, feeds.collection_id, COUNT(items.id) as count
		FROM feeds
		LEFT JOIN items ON items.feed_id = feeds.id AND %s
		WHERE feeds.id IN (%s)
		GROUP BY feeds.id`,
		count_condition, CreatePlaceHolders(len(feed_ids)))

	args := ToAny(feed_ids)
	rows, err := r.readDB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds = []models.FeedResponse{}
	for rows.Next() {
		var feed models.FeedResponse
		err := rows.Scan(&feed.ID, &feed.URL, &feed.Name, &feed.CollectionID, &feed.Count)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}
	return feeds, nil
}

func (r *FeedRepo) UpdateFeed(feed_id int, url, name *string, collection_id *int) error {
	_, err := r.writeDB.Exec(`
        UPDATE feeds SET
			url = COALESCE(?, url),
            name = COALESCE(?, name),
			collection_id = COALESCE(?, collection_id)
        WHERE id = ?
    `, url, name, collection_id, feed_id)
	return err
}

func (r *FeedRepo) RemoveFeedFromCollection(feed_id int) error {
	_, err := r.writeDB.Exec(`
        UPDATE feeds SET collection_id = NULL WHERE id = ?
    `, feed_id)
	return err
}

func (r *FeedRepo) DeleteFeed(feed_id int) error {
	_, err := r.writeDB.Exec(
		"DELETE FROM feeds WHERE id = ?",
		feed_id,
	)
	return err
}
