package models

import (
	"fmt"
	"time"
)

//CreateStatisticsTableQuery query to create statistics table in a SQL database
const CreateStatisticsTableQuery = `CREATE TABLE statistics(
  track_id INTEGER REFERENCES tracks(id),
  listener TEXT,
  listened_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

//CreateListenerIndexQuery listener index
const CreateListenerIndexQuery = `CREATE INDEX IF NOT EXISTS listener_index ON statistics(listener);`

//Statistic tracks listens
type Statistic struct {
	TrackID   int       `json:"-"`
	Listener  string    `json:"listener"`
	Track     *Track    `json:"track"`
	TimeStamp time.Time `json:"timestamp"`
}

//ScanFrom src into stat
func (stat *Statistic) ScanFrom(src scanner) error {
	return src.Scan(&stat.TrackID, &stat.Listener, &stat.TimeStamp)
}

//CreateTable creates tables in db
func (stat *Statistic) CreateTable(db executor) error {
	if _, err := db.Exec(CreateStatisticsTableQuery); err != nil {
		return fmt.Errorf("Error in query: '%s'\nError: %v", CreateStatisticsTableQuery, err)
	}
	if _, err := db.Exec(CreateListenerIndexQuery); err != nil {
		return fmt.Errorf("Error in query: '%s'\nError: %v", CreateListenerIndexQuery, err)
	}
	return nil
}

//CreatePriority order for this entity's create table priority
func (stat *Statistic) CreatePriority() int {
	return 3
}
