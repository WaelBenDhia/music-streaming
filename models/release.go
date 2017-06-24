package models

import "time"

//CreateReleasesTableQuery query to create releases table in a SQL database
var CreateReleasesTableQuery = `CREATE TABLE IF NOT EXISTS releases(
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
func (rel *Release) ScanFrom(src scannable) error {
	return src.Scan(&rel.ID, &rel.ReleaseDate, &rel.Name, &rel.AlbumArtistID, &rel.CoverURL)
}

//Get rel by ID from db
func (rel *Release) Get(db queriable) (bool, error) {
	return notFoundOrErr(rel.ScanFrom(db.QueryRow("SELECT * FROM releases WHERE id = $1;", rel.ID)))
}
