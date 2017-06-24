package models

import (
	"time"
)

//CreateStatisticsTableQuery query to create statistics table in a SQL database
var CreateStatisticsTableQuery = `CREATE TABLE statistics(
  track_id INTEGER REFERENCES tracks(id),
  listener TEXT,
  listened_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);`

//CreateListenerIndexQuery listener index
var CreateListenerIndexQuery = `CREATE INDEX IF NOT EXISTS listener_index ON statistics(listener);`

//Statistic tracks listens
type Statistic struct {
	TrackID   int       `json:"-"`
	Listener  string    `json:"listener"`
	Track     *Track    `json:"track"`
	TimeStamp time.Time `json:"timestamp"`
}

//ScanFrom src into stat
func (stat *Statistic) ScanFrom(src scannable) error {
	return src.Scan(&stat.TrackID, &stat.Listener, &stat.TimeStamp)
}
