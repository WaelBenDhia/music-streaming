package watcher

import (
	"github.com/anacrolix/torrent"
	"github.com/waelbendhia/music-streaming/wms/models"
	mgo "gopkg.in/mgo.v2"
)

//Watcher watches downloading tracks and updates their status accordingly
type Watcher struct {
	activeDownloads []models.Track
	db              *mgo.Database
	torrentCli      *torrent.Client
	stopChannel     chan bool
	downloadDir     string
}

//Start the watcher
func (w *Watcher) Start(db *mgo.Database, torCli *torrent.Client) {
	w.db = db
	w.torrentCli = torCli
	w.stopChannel = make(chan bool, 1)
	go w.watch()
}

//Stop the watcher
func (w *Watcher) Stop() {
	w.stopChannel <- true
}

func (w *Watcher) watch() {
	for {
		select {
		case stop := <-w.stopChannel:
			if stop {
				break
			}
		default:
		}
		// Watch downloads in loop and update DB and torrentCli if download is
		// finished
		for _, tor := range w.torrentCli.Torrents() {
			// Torrent has finished
			if tor.BytesMissing() == 0 {

			}
		}
	}
}
