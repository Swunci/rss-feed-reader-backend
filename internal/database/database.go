package database

import (
	"database/sql"
	"testing"
)

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY)`)
	if err != nil {
		return err
	}

	migrations := []struct {
		version int
		sql     string
	}{
		{1, `CREATE TABLE IF NOT EXISTS feeds (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				url TEXT NOT NULL UNIQUE,
				name TEXT NOT NULL
        	);

			CREATE TABLE IF NOT EXISTS items (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				feed_id INTEGER NOT NULL,
				title TEXT NOT NULL,
				link TEXT NOT NULL,
				description TEXT,
				published_at DATETIME,
				is_read BOOLEAN NOT NULL DEFAULT 0,
				is_favorite BOOLEAN NOT NULL DEFAULT 0,

				FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,

				UNIQUE(feed_id, link)
			);

			CREATE INDEX IF NOT EXISTS idx_items_published_at ON items(published_at);
			
			CREATE TABLE IF NOT EXISTS settings (
				id INTEGER PRIMARY KEY CHECK (id = 1),
				polling_interval INTEGER NOT NULL DEFAULT 30 CHECK (polling_interval >= 5)
			);
    	`},
		{2, `CREATE INDEX IF NOT EXISTS idx_items_read_published 
				ON items (is_read, published_at);
			
			CREATE INDEX IF NOT EXISTS idx_items_favorite_published 
				ON items (is_favorite, published_at
			);
		`},
		{3, `CREATE TABLE collections (
  				id   INTEGER PRIMARY KEY AUTOINCREMENT,
  				name TEXT NOT NULL UNIQUE
			);

			ALTER TABLE feeds
  				ADD COLUMN collection_id INTEGER REFERENCES collections(id) ON DELETE SET NULL;
		`},
	}

	for _, m := range migrations {
		var count int
		db.QueryRow("SELECT COUNT(*) FROM schema_version WHERE version = ?", m.version).Scan(&count)
		if count > 0 {
			continue
		}
		if _, err := db.Exec(m.sql); err != nil {
			return err
		}
		db.Exec("INSERT INTO schema_version (version) VALUES (?)", m.version)
	}
	return nil
}

func SetupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.Exec("PRAGMA foreign_keys = ON")
	table_err := Migrate(db)
	if table_err != nil {
		t.Fatal(table_err)
	}
	return db
}
