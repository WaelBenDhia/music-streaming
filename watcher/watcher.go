package watcher

import (
	"github.com/anacrolix/torrent"
	"github.com/wael/music-streaming/wms/models"
	mgo "gopkg.in/mgo.v2"
)

//Watcher watches downloading tracks and updates their status accordingly
type Watcher struct {
	activeDownloads []models.Track
	db              *mgo.Database
	torrentCli      *torrent.Client
	stopChannel     chan bool
}

func (w *Watcher) Start(db *mgo.Database, torCli *torrent.Client) {
	w.db = db
	w.torrentCli = torCli
	w.stopChannel = make(chan bool, 1)
	go w.watch()
}

func (w *Watcher) watch() {
	run := true
	for run {
		select {
		case stop := <-w.stopChannel:
			if stop {
				break
			}
		default:
		}
		//Watch downloads in loop and update DB and torrentCli if download is finished
	}
}
