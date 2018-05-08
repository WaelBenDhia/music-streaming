package server

import (
	"net/http"
	"time"

	"encoding/json"

	"log"

	"github.com/waelbendhia/music-streaming/gopirate"
	"github.com/waelbendhia/music-streaming/wms/models"
)

func (s *Server) searchAlbumsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		name        = r.URL.Query().Get("name")
		search      = models.Release{Name: name}
		finalResult []models.Release
		intResC     = make(chan []models.Release, 1)
		lfmResC     = make(chan []models.Release, 1)
		intErrC     = make(chan error, 1)
		lfmErrC     = make(chan error, 1)
		run         = func(
			resC chan<- []models.Release,
			errC chan<- error,
			fn func() ([]models.Release, error),
		) {
			res, err := fn()
			resC <- res
			errC <- err
			close(resC)
			close(errC)
		}
	)
	go run(intResC, intErrC, func() ([]models.Release, error) {
		return search.Search(s.db)
	})
	go run(lfmResC, lfmErrC, func() ([]models.Release, error) {
		return lfmSearchConverter(s.lfmCli.SearchAlbums(name))
	})
	wait := func(
		resC <-chan []models.Release,
		errC <-chan error,
	) (res []models.Release, err error) {
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
	fmAlbum, err := s.lfmCli.GetAlbumInfo(album.AlbumArtist.Name, album.Name)
	panicIfErr(err)
	searchString := album.Name
	if album.AlbumArtist != nil {
		searchString = album.AlbumArtist.Name + " " + album.Name
	}
	res, err := gopirate.Search(searchString)
	if err == gopirate.ErrNoResults {
		http.Error(w, "no results found", 404)
		return
	}
	converted := lfmAlbumConverter(&fmAlbum.Album)
	converted.Save(s.db)
	res = torrentSort(res, func(tor gopirate.Torrent) int {
		score := scoreTorrentHealth(tor) + scoreTorrentName(searchString)(tor)
		return score
	})
	s.torrentCli.AddTPBTorrent(res[0])
	w.WriteHeader(200)
	output, err := json.Marshal(res)
	panicIfErr(err)
	w.WriteHeader(200)
	panicIfErr(w.Write(output))
	go func() {
		_ = <-s.torrentCli.GotInfo(res[0])
		s.infoLog.Println(s.torrentCli.GetInfo(res[0]).Files)
		matched := matchTracksToFiles(
			converted.Tracks,
			s.torrentCli.GetTorrent(res[0]).Files(),
		)
		log.Println(converted.Tracks)
		for _, v := range matched {
			v.Download()
			log.Println("Downloading", v.DisplayPath())
		}
		log.Println(matched)
		// panicIfErr(s.torrentCli.StartAll(res[0]))
		go func() {
			var prev int64
			for {
				s.torrentCli.PrintStatus(res[0], s.infoLog)
				now := s.torrentCli.GetTorrent(res[0]).BytesCompleted()
				delta := (now - prev) / (5 * 1024)
				prev = now
				s.infoLog.Println("Delta", delta, "kbps")
				s.torrentCli.GetTorrent(res[0]).BytesCompleted()
				time.Sleep(5 * time.Second)
			}
		}()
	}()
}
