package repositories

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Swunci/rss-feed-backend/internal/models"
)

type ItemRepo struct {
	readDB  *sql.DB
	writeDB *sql.DB
}

func NewItemRepo(readDB *sql.DB, writeDB *sql.DB) *ItemRepo {
	return &ItemRepo{readDB: readDB, writeDB: writeDB}
}

func (r *ItemRepo) CreateItems(feed_id int, items []models.Item) error {
	if len(items) == 0 {
		return nil
	}

	placeholders := make([]string, len(items))
	args := make([]any, 0, len(items)*5)

	for i, item := range items {
		placeholders[i] = "(?, ?, ?, ?, ?)"
		args = append(args, item.FeedID, item.Title, item.Link, item.Description, item.PublishedAt.UTC().Format(time.RFC3339))
	}

	query := fmt.Sprintf(`INSERT OR IGNORE INTO items (feed_id, title, link, description, published_at)
			VALUES %s`, strings.Join(placeholders, ","),
	)
	res, err := r.writeDB.Exec(query, args...)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	slog.Debug("Items created", "count", count)
	return err

}

func (r *ItemRepo) GetItem(item_id int) (models.Item, error) {
	row := r.readDB.QueryRow(
		`SELECT * FROM items where id = ?`, item_id,
	)
	var item models.Item

	err := row.Scan(
		&item.ID,
		&item.FeedID,
		&item.Title,
		&item.Link,
		&item.Description,
		&item.PublishedAt,
		&item.IsRead,
		&item.IsFavorite,
	)
	if err != nil {
		return models.Item{}, err
	}
	return item, nil
}

func (r *ItemRepo) GetAllItems(filter models.ItemFilter, timestamp_cursor string, limit int) ([]models.Item, error) {
	query := `SELECT * FROM items WHERE 1=1`
	args := []any{}
	query, args, err := ApplyItemFilters(query, args, filter, timestamp_cursor)
	if err != nil {
		return nil, err
	}
	query += ` ORDER BY published_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}

	rows, err := r.readDB.Query(query, args...)
	if err != nil {
		return nil, err
	}

	var items = []models.Item{}

	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID,
			&item.FeedID,
			&item.Title,
			&item.Link,
			&item.Description,
			&item.PublishedAt,
			&item.IsRead,
			&item.IsFavorite,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *ItemRepo) GetItemsByFeed(feed_id int, filter models.ItemFilter, timestamp_cursor string, limit int) ([]models.Item, error) {
	query := `SELECT * FROM items WHERE feed_id = ?`
	args := []any{feed_id}
	query, args, err := ApplyItemFilters(query, args, filter, timestamp_cursor)
	if err != nil {
		return nil, err
	}

	query += ` ORDER BY published_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}
	rows, err := r.readDB.Query(query, args...)
	if err != nil {
		return nil, err
	}

	var items = []models.Item{}

	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID,
			&item.FeedID,
			&item.Title,
			&item.Link,
			&item.Description,
			&item.PublishedAt,
			&item.IsRead,
			&item.IsFavorite,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *ItemRepo) GetItemsByCollection(collection_id int, filter models.ItemFilter, timestamp_cursor string, limit int) ([]models.Item, error) {
	query := `SELECT items.* FROM items 
			  JOIN feeds ON items.feed_id = feeds.id 
			  WHERE feeds.collection_id = ?`
	args := []any{collection_id}
	query, args, err := ApplyItemFilters(query, args, filter, timestamp_cursor)
	if err != nil {
		return nil, err
	}

	query += ` ORDER BY published_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}
	rows, err := r.readDB.Query(query, args...)
	if err != nil {
		return nil, err
	}

	var items = []models.Item{}

	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID,
			&item.FeedID,
			&item.Title,
			&item.Link,
			&item.Description,
			&item.PublishedAt,
			&item.IsRead,
			&item.IsFavorite,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *ItemRepo) GetUnreadItemsFeedIds() ([]int, error) {
	rows, err := r.readDB.Query(
		`SELECT DISTINCT feed_id FROM items WHERE is_read = 0`,
	)
	if err != nil {
		return nil, err
	}

	var feed_ids = []int{}

	for rows.Next() {
		var feed_id int
		err := rows.Scan(&feed_id)
		if err != nil {
			return nil, err
		}
		feed_ids = append(feed_ids, feed_id)
	}
	return feed_ids, nil
}

func (r *ItemRepo) GetFavoriteItemsFeedIds() ([]int, error) {
	rows, err := r.readDB.Query(
		`SELECT DISTINCT feed_id FROM items WHERE is_favorite = 1`,
	)
	if err != nil {
		return nil, err
	}

	var feed_ids = []int{}

	for rows.Next() {
		var feed_id int
		err := rows.Scan(&feed_id)
		if err != nil {
			return nil, err
		}
		feed_ids = append(feed_ids, feed_id)
	}
	return feed_ids, nil
}

func (r *ItemRepo) UpdateRead(item_id int, is_read bool) error {
	_, err := r.writeDB.Exec(
		"UPDATE items SET is_read = ? WHERE id = ?", is_read, item_id,
	)
	slog.Debug("Marked as read/unread", "id", item_id, "is_read", is_read)
	return err
}

func (r *ItemRepo) UpdateReadMultiple(item_ids []int, is_read bool) error {
	placeholders := make([]string, len(item_ids))
	args := make([]any, len(item_ids)+1)
	args[0] = is_read
	for i, id := range item_ids {
		placeholders[i] = "?"
		args[i+1] = id
	}

	query := fmt.Sprintf(
		"UPDATE items SET is_read = ? WHERE id IN (%s)",
		strings.Join(placeholders, ","),
	)

	_, err := r.writeDB.Exec(query, args...)
	return err
}

func (r *ItemRepo) UpdateFavorite(item_id int, is_favorite bool) error {
	_, err := r.writeDB.Exec(
		"UPDATE items SET is_favorite = ? WHERE id = ?", is_favorite, item_id,
	)
	slog.Debug("Favorite/unfavorite", "id", item_id, "is_favorite", is_favorite)
	return err
}

func (r *ItemRepo) DeleteItem(item_id int) error {
	_, err := r.writeDB.Exec(
		"DELETE FROM items WHERE id = ?",
		item_id,
	)
	return err
}
