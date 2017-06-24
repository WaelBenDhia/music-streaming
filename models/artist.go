package models

//CreateArtistsTableQuery query to create artists table in a SQL database
var CreateArtistsTableQuery = `CREATE TABLE IF NOT EXISTS artists(
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  image_url TEXT
);`

//CreateArtistRelationsTableQuery query to create artist relations table in a SQL database
var CreateArtistRelationsTableQuery = `CREATE TABLE IF NOT EXISTS artist_relations(
  id1 INTEGER REFERENCES artists(id),
  id2 INTEGER REFERENCES artists(id)
);`

//Artist represents an artist/band/person
type Artist struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	ImageURL       string   `json:"imageURL"`
	RelatedArtists []Artist `json:"relatedArtists,omitempty"`
}

//ScanFrom src into artist
func (artist *Artist) ScanFrom(src scannable) error {
	return src.Scan(&artist.ID, &artist.Name, &artist.ImageURL)
}

//Get artist by ID from db
func (artist *Artist) Get(db queriable) (bool, error) {
	return notFoundOrErr(artist.ScanFrom(db.QueryRow("SELECT * FROM artists WHERE id = $1;", artist.ID)))
}
