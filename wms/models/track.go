package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const trackColName = "track"

//Track represents an artist/band/person
type Track struct {
	ID       bson.ObjectId `json:"id" bson:"_id"`
	Name     string        `json:"name" bson:"name"`
	Length   time.Duration `json:"length" bson:"length"`
	TrackURL string        `json:"-" bson:"track_url"`
	Genre    string        `json:"genre" bson:"genre"`
	ArtistID int           `json:"-" bson:"artist_id"`
	Artist   *Artist       `json:"artist,omitempty" bson:"-"`
	Releases []Release     `json:"releases,omitempty" bson:"-"`
}

//Get track by ID from db
func (track *Track) Get(db *mgo.Database) (bool, error) {
	return notFoundOrErr(db.C(trackColName).Find(bson.M{"_id": track.ID}).One(track))
}

//Search for tracks by
func (track *Track) Search(db *mgo.Database) ([]Track, error) {
	var tracks []Track
	err := db.C(trackColName).Find(bson.M{"artist_id": track.ArtistID}).One(&tracks)
	return tracks, err
}

//ColCreate creates a collection in db with the appropriate indexes
func (track *Track) ColCreate(db *mgo.Database) error {
	return db.C(trackColName).EnsureIndex(mgo.Index{Key: []string{"track_url"}})
}

//Save rel to db
func (track *Track) Save(db *mgo.Database) error {
	track.ID = bson.NewObjectId()
	return db.C(relColName).Insert(track)
}
