package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func ParseID(id string) (int, error) {
	int_id, err := strconv.Atoi(id)
	if err != nil {
		return 0, err
	}
	return int_id, nil
}
