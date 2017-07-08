package models

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const relColName = "release"

//Release represents an artist/band/person
type Release struct {
	ID            bson.ObjectId `json:"id,omitempty" bson:"_id"`
	ReleaseDate   time.Time     `json:"releaseDate,omitempty" bson:"release_date"`
	Name          string        `json:"name,omitempty" bson:"name"`
	AlbumArtistID string        `json:"-" bson:"album_artist_id"`
	AlbumArtist   *Artist       `json:"artist,omitempty" bson:"-"`
	CoverURL      string        `json:"coverURL,omitempty" bson:"cover_url"`
	TrackIDs      []string      `json:"-" bson:"track_ids"`
	Tracks        []Track       `json:"tracks,omitempty" bson:"-"`
}

//Get rel by ID or Name from db
func (rel *Release) Get(db *mgo.Database) (bool, error) {
	finder := bson.M{"name": rel.Name}
	if rel.Name == "" {
		finder = bson.M{"_id": rel.ID}
	}
	return notFoundOrErr(db.C(relColName).Find(finder).One(rel))
}

//GetFull rel by ID from db
func (rel *Release) GetFull(db *mgo.Database) (bool, error) {
	found, err := rel.Get(db)
	if !found || err != nil {
		return found, err
	}
	rel.AlbumArtist = &Artist{ID: bson.ObjectId(rel.AlbumArtistID)}
	found, err = rel.AlbumArtist.Get(db)
	if !found || err != nil {
		return found, err
	}
	var quErr error
	for _, trcID := range rel.TrackIDs {
		track := Track{ID: bson.ObjectId(trcID)}
		found, quErr = track.Get(db)
		if !found {
			quErr = fmt.Errorf("Track with ID: '%s' not found", string(track.ID))
		}
		if quErr != nil {
			break
		}
	}
	return true, quErr
}

//ColCreate creates tables in db
func (rel *Release) ColCreate(db *mgo.Database) error {
	return db.C(relColName).EnsureIndex(mgo.Index{Key: []string{"name", "album_artist_id"}, Unique: true})
}

//Search for releases by artist then name
func (rel *Release) Search(db *mgo.Database) ([]Release, error) {
	var (
		rels []Release
		err  error
	)
	if rel.AlbumArtistID != "" {
		err = db.C(relColName).Find(bson.M{"album_artist_id": rel.AlbumArtistID}).All(&rels)
	} else {
		err = db.C(relColName).Find(bson.M{"name": rel.Name}).All(&rels)
	}
	return rels, err
}

//Save rel to db
func (rel *Release) Save(db *mgo.Database) error {
	rel.ID = bson.NewObjectId()
	return db.C(relColName).Insert(rel)
}
