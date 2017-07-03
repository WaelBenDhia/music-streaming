package models

import (
	"time"

	"gopkg.in/mgo.v2"
)

const statColName = "statistics"

//Statistic tracks listens
type Statistic struct {
	TrackID   string    `json:"-" bson:"track_id"`
	Listener  string    `json:"listener" bson:"listener_ip"`
	Track     *Track    `json:"track" bson:"-"`
	TimeStamp time.Time `json:"timestamp" bson:"timestamp"`
}

//Save rel to db
func (stat *Statistic) Save(db *mgo.Database) error {
	return db.C(relColName).Insert(stat)
}

//ColCreate creates collection in db
func (stat *Statistic) ColCreate(db *mgo.Database) error {
	return nil
}
