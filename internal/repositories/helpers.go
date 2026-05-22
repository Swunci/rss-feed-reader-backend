package repositories

import "strings"

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
