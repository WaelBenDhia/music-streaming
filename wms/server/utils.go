package server

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/texttheater/golang-levenshtein/levenshtein"
	"github.com/wael/music-streaming/gopirate"
	"github.com/wael/music-streaming/lastfm"
	"github.com/wael/music-streaming/wms/models"
)

//AddMiddleware creates a new handler adapted with middleware
func AddMiddleware(h http.HandlerFunc) func(...middleware) http.Handler {
	return func(ads ...middleware) http.Handler {
		var handler http.Handler
		handler = h
		for _, mw := range ads {
			handler = mw(handler)
		}
		return handler
	}
}

func ctxWithValCancel(ctx context.Context, valKey key, val interface{}) (context.Context, context.CancelFunc) {
	return context.WithCancel(context.WithValue(ctx, valKey, val))
}

func lfmSearchConverter(lfmSearch *lastfm.SearchResults, err error) ([]models.Release, error) {
	if err != nil {
		return nil, err
	}
	res := make([]models.Release, len(lfmSearch.Results.AlbumMatches.Album))
	for i, alb := range lfmSearch.Results.AlbumMatches.Album {
		res[i] = lfmAlbumConverter(&alb)
	}
	return res, nil
}

func lfmAlbumConverter(lfmAlbum *lastfm.Album) models.Release {
	var album models.Release
	album.AlbumArtist = &models.Artist{Name: lfmAlbum.Artist}
	album.Name = lfmAlbum.Name
	for _, image := range lfmAlbum.Image {
		if image.Size == "extralarge" {
			album.CoverURL = image.Text
			break
		}
	}
	for _, track := range lfmAlbum.Tracks.Tracks {
		var newTrack models.Track
		duration, convErr := strconv.Atoi(track.Duration)
		if convErr != nil {
			duration = 0
		}
		newTrack.Length = time.Duration(duration) * time.Second
		newTrack.Name = track.Name
		album.Tracks = append(album.Tracks, newTrack)
	}
	return album
}

func panicIfErr(args ...interface{}) {
	if err, ok := args[len(args)-1].(error); ok && err != nil {
		panic(err)
	}
}

func scoreTorrentName(name string) func(gopirate.Torrent) int {
	return func(tor gopirate.Torrent) int {
		return levenshtein.DistanceForStrings([]rune(name), []rune(tor.Name), levenshtein.DefaultOptions)
	}
}

func scoreTorrentHealth(tor gopirate.Torrent) int {
	return tor.Seeders - tor.Leechers
}

func torrentSort(a []gopirate.Torrent, score func(gopirate.Torrent) int) []gopirate.Torrent {
	if len(a) < 2 {
		return a
	}
	left, right, pivot := 0, len(a)-1, rand.Int()%len(a)
	a[pivot], a[right] = a[right], a[pivot]
	for i := range a {
		if score(a[i]) < score(a[right]) {
			a[i], a[left] = a[left], a[i]
			left++
		}
	}
	a[left], a[right] = a[right], a[left]
	torrentSort(a[:left], score)
	torrentSort(a[left+1:], score)
	return a
}
