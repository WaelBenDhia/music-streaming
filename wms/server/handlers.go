package server

import (
	"net/http"
	"time"

	"encoding/json"

	"log"

	"github.com/wael/music-streaming/gopirate"
	"github.com/wael/music-streaming/wms/models"
)

func (s *Server) searchAlbumsHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	search := models.Release{Name: name}
	var finalResult []models.Release
	intResC, lfmResC := make(chan []models.Release, 1), make(chan []models.Release, 1)
	intErrC, lfmErrC := make(chan error, 1), make(chan error, 1)
	run := func(resC chan<- []models.Release, errC chan<- error, fn func() ([]models.Release, error)) {
		res, err := fn()
		resC <- res
		errC <- err
		close(resC)
		close(errC)
	}
	go run(intResC, intErrC, func() ([]models.Release, error) { return search.Search(s.db) })
	go run(lfmResC, lfmErrC, func() ([]models.Release, error) { return lfmSearchConverter(s.lfmCli.SearchAlbums(name)) })
	wait := func(resC <-chan []models.Release, errC <-chan error) (res []models.Release, err error) {
		select {
		case res = <-resC:
		case err = <-errC:
		}
		return res, err
	}
	intRes, intErr := wait(intResC, intErrC)
	lfmRes, lfmErr := wait(lfmResC, lfmErrC)
	panicIfErr(intErr)
	panicIfErr(lfmErr)
	finalResult = append(intRes, lfmRes...)
	output, err := json.Marshal(finalResult)
	panicIfErr(err)
	w.WriteHeader(200)
	panicIfErr(w.Write(output))
}

func (s *Server) downloadAlbumHandler(w http.ResponseWriter, r *http.Request) {
	album := r.Context().Value(requestKey).(*models.Release)
	searchString := album.Name
	if album.AlbumArtist != nil {
		searchString = album.AlbumArtist.Name + " " + album.Name
	}
	res, err := gopirate.Search(searchString)
	if err == gopirate.ErrNoResults {
		http.Error(w, "no results found", 404)
		return
	}
	s.lfmCli.GetAlbumInfo(album.AlbumArtist.Name, album.Name)
	panicIfErr(err)
	res = torrentSort(res, func(tor gopirate.Torrent) int {
		score := scoreTorrentHealth(tor) + scoreTorrentName(searchString)(tor)
		log.Println("Torrent: ", tor.Name, " Score: ", score)
		return score
	})
	log.Println("SELECTED: ", res[0])
	s.torrentCli.AddTPBTorrent(res[0])
	w.WriteHeader(200)
	output, err := json.Marshal(res)
	panicIfErr(err)
	w.WriteHeader(200)
	panicIfErr(w.Write(output))
	go func() {
		_ = <-s.torrentCli.GotInfo(res[0])
		s.infoLog.Println(s.torrentCli.GetInfo(res[0]).Files)
		panicIfErr(s.torrentCli.StartAll(res[0]))
		go func() {
			for {
				s.torrentCli.PrintStatus(res[0], s.infoLog)
				time.Sleep(5 * time.Second)
			}
		}()
	}()
}
