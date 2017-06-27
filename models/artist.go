package models

import (
	"database/sql"
	"fmt"
)

//CreateArtistsTableQuery query to create artists table in a SQL database
const CreateArtistsTableQuery = `CREATE TABLE IF NOT EXISTS artists(
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  image_url TEXT
);`

//CreateArtistRelationsTableQuery query to create artist relations table in a SQL database
const CreateArtistRelationsTableQuery = `CREATE TABLE IF NOT EXISTS artist_relations(
  id1 INTEGER REFERENCES artists(id),
  id2 INTEGER REFERENCES artists(id)
);`

//Artist represents an artist/band/person
type Artist struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	ImageURL       string    `json:"imageURL"`
	RelatedArtists []Artist  `json:"relatedArtists,omitempty"`
	Releases       []Release `json:"releases,omitempty"`
}

//ScanFrom src into artist
func (artist *Artist) ScanFrom(src scanner) error {
	return src.Scan(&artist.ID, &artist.Name, &artist.ImageURL)
}

//Get artist by ID from db
func (artist *Artist) Get(db querier) (bool, error) {
	return notFoundOrErr(artist.ScanFrom(db.QueryRow("SELECT * FROM artists WHERE id = $1;", artist.ID)))
}

//GetFull artist by ID from db
func (artist *Artist) GetFull(db querier) (bool, error) {
	if found, err := artist.Get(db); !found || err != nil {
		return found, err
	}
	rel := Release{AlbumArtistID: artist.ID}
	var err error
	artist.Releases, err = rel.Search(db)
	if err != nil {
		return true, err
	}
	var rows *sql.Rows
	rows, err = db.Query(`SELECT art.* FROM artists art INNER JOIN (SELECT a.id1 as id FROM artists_relations WHERE id2 = ? UNION SELECT b.id2 as id FROM artist_relations WHERE id1 = ?) rels on art.id = rels.id;`, artist.ID, artist.ID)
	defer func() {
		if rows != nil {
			err = errOr(err, rows.Close())
		}
	}()
	for rows.Next() && err == nil {
		var relArt Artist
		err = relArt.ScanFrom(rows)
		artist.RelatedArtists = append(artist.RelatedArtists, relArt)
	}
	return true, err
}

//CreateTable creates tables in db
func (artist *Artist) CreateTable(db executor) error {
	if _, err := db.Exec(CreateArtistsTableQuery); err != nil {
		return fmt.Errorf("Error in query: '%s'\nError: %v", CreateArtistsTableQuery, err)
	}
	if _, err := db.Exec(CreateArtistRelationsTableQuery); err != nil {
		return fmt.Errorf("Error in query: '%s'\nError: %v", CreateArtistRelationsTableQuery, err)
	}

	return nil
}

//CreatePriority order for this entity's create table priority
func (artist *Artist) CreatePriority() int {
	return 0
}
