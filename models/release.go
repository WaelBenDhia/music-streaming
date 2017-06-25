package models

import (
	"fmt"
	"time"
)

//CreateReleasesTableQuery query to create releases table in a SQL database
const CreateReleasesTableQuery = `CREATE TABLE IF NOT EXISTS releases(
  id SERIAL PRIMARY KEY,
  release_date DATE,
  name TEXT NOT NULL,
  album_artist_id INTEGER REFERENCES artists(id),
  cover_url TEXT
);`

//Release represents an artist/band/person
type Release struct {
	ID            int       `json:"id"`
	ReleaseDate   time.Time `json:"releaseDate"`
	Name          string    `json:"name"`
	AlbumArtistID int       `json:"-"`
	AlbumArtist   *Artist   `json:"artist,omitempty"`
	CoverURL      string    `json:"coverURL"`
	Tracks        []Track   `json:"tracks"`
}

//ScanFrom src into rel
func (rel *Release) ScanFrom(src scanner) error {
	return src.Scan(&rel.ID, &rel.ReleaseDate, &rel.Name, &rel.AlbumArtistID, &rel.CoverURL)
}

//Get rel by ID from db
func (rel *Release) Get(db querier) (bool, error) {
	return notFoundOrErr(rel.ScanFrom(db.QueryRow("SELECT * FROM releases WHERE id = $1;", rel.ID)))
}

//CreateTable creates tables in db
func (rel *Release) CreateTable(db executor) error {
	if _, err := db.Exec(CreateReleasesTableQuery); err != nil {
		return fmt.Errorf("Error in query: '%s'\nError: %v", CreateReleasesTableQuery, err)
	}
	return nil
}

//CreatePriority order for this entity's create table priority
func (rel *Release) CreatePriority() int {
	return 1
}
