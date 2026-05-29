package repositories

import (
	"fmt"
	"strings"
	"time"

	"github.com/Swunci/rss-feed-backend/internal/models"
)

func CreatePlaceHolders(size int) string {
	placeholders := make([]string, size)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ",")
}

func ToAny[T any](slice []T) []any {
	args := make([]any, len(slice))
	for i, v := range slice {
		args[i] = v
	}
	return args
}

func ApplyItemFilters(query string, args []any, filter models.ItemFilter, timestampCursor string) (string, []any, error) {
	if filter.IsRead != nil {
		query += ` AND is_read = ?`
		args = append(args, *filter.IsRead)
	}
	if filter.IsFavorite != nil {
		query += ` AND is_favorite = ?`
		args = append(args, *filter.IsFavorite)
	}
	if timestampCursor != "" {
		timestamp, err := time.Parse(time.RFC3339, timestampCursor)
		if err != nil {
			return query, args, fmt.Errorf("invalid timestamp cursor: %w", err)
		}
		query += ` AND published_at < ?`
		args = append(args, timestamp.UTC().Format(time.RFC3339))
	}
	return query, args, nil
}
