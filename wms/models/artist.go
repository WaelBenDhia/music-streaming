package models

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const artistColName = "artist"

//Artist represents an artist/band/person
type Artist struct {
	ID               bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Name             string        `json:"name,omitempty" bson:"name"`
	ImageURL         string        `json:"imageURL,omitempty" bson:"image_url"`
	RelatedArtistIDs []string      `json:"-" bson:"related_artist_ids"`
	RelatedArtists   []Artist      `json:"relatedArtists,omitempty" bson:"-"`
	Releases         []Release     `json:"releases,omitempty" bson:"-"`
}

//Get artist by ID from db
func (artist *Artist) Get(db *mgo.Database) (bool, error) {
	finder := bson.M{"name": artist.Name}
	if artist.Name == "" {
		finder = bson.M{"_id": artist.ID}
	}
	return notFoundOrErr(db.C(artistColName).Find(finder).One(artist))
}

//GetFull artist by ID from db
func (artist *Artist) GetFull(db *mgo.Database) (bool, error) {
	if found, err := artist.Get(db); !found || err != nil {
		return found, err
	}
	rel := Release{AlbumArtistID: string(artist.ID)}
	var err error
	artist.Releases, err = rel.Search(db)
	if err != nil {
		return true, err
	}
	var quErr error
	for _, relID := range artist.RelatedArtistIDs {
		relArt := Artist{ID: bson.ObjectId(relID)}
		var found bool
		found, quErr = relArt.Get(db)
		if !found {
			quErr = fmt.Errorf("Artist with ID: '%s' not found", string(relArt.ID))
		}
		if quErr != nil {
			break
		}
	}
	return true, quErr
}

//Save artist into db
func (artist *Artist) Save(db *mgo.Database) error {
	artist.ID = bson.NewObjectId()
	return db.C(artistColName).Insert(artist)
}

//ColCreate creates a collection in db with the appropriate indexes
func (artist *Artist) ColCreate(db *mgo.Database) error {
	return db.C(artistColName).EnsureIndex(mgo.Index{Key: []string{"name"}, Unique: true})
}
