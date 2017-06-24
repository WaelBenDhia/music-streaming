package models

import "time"

//CreateTracksTableQuery query to create tracks table in a SQL database
var CreateTracksTableQuery = `CREATE TABLE IF NOT EXISTS tracks(
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  length INTERVAL,
  track_url TEXT,
  genre TEXT,
  artist_id INTEGER REFERENCES artists(id)
);`

//CreateGenreIndexQuery query to create an index on the genre column
var CreateGenreIndexQuery = `CREATE INDEX IF NOT EXISTS genre_index ON tracks(genre);`

//CreateTrackReleaseRelationTableQuery query to create release<=>track relation table in a SQL database
var CreateTrackReleaseRelationTableQuery = `CREATE TABLE IF NOT EXISTS track_release_relations(
  track_id INTEGER REFERENCES tracks(id),
  release_id INTEGER REFERENCES releases(id)
);`

//Track represents an artist/band/person
type Track struct {
	ID       int           `json:"id"`
	Name     string        `json:"name"`
	Length   time.Duration `json:"length"`
	TrackURL string        `json:"-"`
	Genre    string        `json:"genre"`
	ArtistID int           `json:"-"`
	Artist   *Artist       `json:"artist,omitempty"`
	Releases []Release     `json:"releases,omitempty"`
}

//ScanFrom src into track
func (track *Track) ScanFrom(src scannable) error {
	return src.Scan(&track.ID, &track.Name, &track.Length, &track.TrackURL, &track.Genre, &track.ArtistID)
}

//Get track by ID from db
func (track *Track) Get(db queriable) (bool, error) {
	return notFoundOrErr(track.ScanFrom(db.QueryRow("SELECT * FROM tracks WHERE id = $1;", track.ID)))
}
