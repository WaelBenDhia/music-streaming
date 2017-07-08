package server

import (
	"net/http"

	"encoding/json"

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
	res, err := gopirate.Search(album.AlbumArtist.Name + " " + album.Name)
	panicIfErr(err)
	w.WriteHeader(200)
	output, err := json.Marshal(res)
	panicIfErr(err)
	w.WriteHeader(200)
	panicIfErr(w.Write(output))
}
