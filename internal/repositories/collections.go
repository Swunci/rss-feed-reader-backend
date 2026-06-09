package repositories

import (
	"database/sql"

	"github.com/Swunci/rss-feed-backend/internal/models"
)

type CollectionRepo struct {
	readDB  *sql.DB
	writeDB *sql.DB
}

func NewCollectionRepo(readDB *sql.DB, writeDB *sql.DB) *CollectionRepo {
	return &CollectionRepo{readDB: readDB, writeDB: writeDB}
}

func (r *CollectionRepo) CreateCollection(name string) (models.Collection, error) {
	_, err := r.writeDB.Exec("INSERT OR IGNORE INTO collections (name) VALUES (?)", name)
	if err != nil {
		return models.Collection{}, err
	}

	collection := models.Collection{}
	err = r.readDB.QueryRow("SELECT * FROM collections where name = ?", name).
		Scan(&collection.ID, &collection.Name)

	return collection, err
}

func (r *CollectionRepo) GetCollections() ([]models.Collection, error) {
	rows, err := r.readDB.Query("SELECT * from collections")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	collections := []models.Collection{}

	for rows.Next() {
		collection := models.Collection{}
		err = rows.Scan(&collection.ID, &collection.Name)
		if err != nil {
			return []models.Collection{}, err
		}
		collections = append(collections, collection)
	}
	return collections, nil
}

func (r *CollectionRepo) UpdateCollection(collection_id int, name string) error {
	_, err := r.writeDB.Exec("UPDATE collections SET name = (?) WHERE id = ?", name, collection_id)
	return err
}

func (r *CollectionRepo) DeleteCollection(collection_id int) error {
	_, err := r.writeDB.Exec("DELETE from collections where id = ?", collection_id)
	return err
}
