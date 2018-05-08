package server

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	"github.com/waelbendhia/music-streaming/gopirate"
	"github.com/waelbendhia/music-streaming/lastfm"
	"github.com/waelbendhia/music-streaming/wms/models"
	"gopkg.in/mgo.v2/bson"
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

func ctxWithValCancel(
	ctx context.Context,
	valKey key,
	val interface{},
) (context.Context, context.CancelFunc) {
	return context.WithCancel(context.WithValue(ctx, valKey, val))
}

func lfmSearchConverter(
	lfmSearch *lastfm.SearchResults,
	err error,
) ([]models.Release, error) {
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
		return levDistance(name, tor.Name)
	}
}

func levDistance(first, second string) int {
	return levenshtein.DistanceForStrings(
		[]rune(first),
		[]rune(second),
		levenshtein.DefaultOptions,
	)
}

func scoreTorrentHealth(tor gopirate.Torrent) int {
	if tor.Seeders == 0 {
		return math.MaxInt32
	}
	return tor.Leechers - tor.Seeders
}

func torrentSort(
	a []gopirate.Torrent,
	score func(gopirate.Torrent) int,
) []gopirate.Torrent {
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
func lfmAlbumInfoWrapper(
	lfmAlb *lastfm.Album,
	err error,
) (*models.Release, error) {
	if err != nil {
		return nil, err
	}
	rel := &models.Release{
		Name: lfmAlb.Name,
		AlbumArtist: &models.Artist{
			Name: lfmAlb.Artist,
		},
	}
	for _, track := range lfmAlb.Tracks.Tracks {
		rel.Tracks = append(rel.Tracks, models.Track{
			Name: track.Name,
		})
	}
	return rel, err
}

// type recResult struct {
// 	pairs map[bson.ObjectId]torrent.File
// 	dist  int
// }

// func hashSlice(tracks []models.Track, files []torrent.File) string {
// 	hasher := md5.New()
// 	hasher.Write([]byte{byte(len(tracks))})
// 	for _, file := range files {
// 		_, _ = hasher.Write([]byte(file.DisplayPath()))
// 	}
// 	return string(hasher.Sum(nil))
// }

// func recurse(tracks []models.Track, files []torrent.File, memo map[string]recResult) (map[bson.ObjectId]torrent.File, int) {
// 	//Stop condition
// 	if len(tracks) < 1 {
// 		return make(map[bson.ObjectId]torrent.File), 0
// 	}
// 	id := hashSlice(tracks, files)
// 	//Check memo
// 	if prev, ok := memo[id]; ok {
// 		return prev.pairs, prev.dist
// 	}
// 	var (
// 		bestRes recResult
// 		bestI   int
// 	)
// 	bestRes.dist = math.MaxInt32
// 	for i := range files {
// 		individualDist := levDistance(
// 			strings.ToLower(tracks[0].Name),
// 			strings.ToLower(filepath.Base(files[i].DisplayPath())))
// 		res, dist := recurse(tracks[1:], append(
// 			append([]torrent.File{}, files[:i]...),
// 			files[i+1:]...), memo)
// 		dist += individualDist
// 		if dist < bestRes.dist {
// 			bestI = i
// 			bestRes = recResult{res, dist}
// 		}
// 	}
// 	bestRes.pairs[tracks[0].ID] = files[bestI]
// 	memo[id] = bestRes
// 	return bestRes.pairs, bestRes.dist
// }

// func matchTracksToFiles(tracks []models.Track, files []torrent.File) map[bson.ObjectId]torrent.File {
// 	log.Println("::::::::::::::::::::::Matching:::::::::::::::::::::::::")
// 	var memo = make(map[string]recResult)
// 	res, dist := recurse(tracks, files, memo)
// 	log.Println("Best Dist", dist)
// 	log.Println("Size map", len(res))
// 	findTrack := func(id bson.ObjectId) models.Track {
// 		for _, track := range tracks {
// 			if track.ID == id {
// 				return track
// 			}
// 		}
// 		return models.Track{}
// 	}
// 	for k, v := range res {
// 		log.Println("Matched", findTrack(k).Name, "to", filepath.Base(v.DisplayPath()))
// 	}
// 	return res
// }

func matchTracksToFiles(
	tracks []models.Track,
	files []torrent.File,
) map[bson.ObjectId]*torrent.File {
	var tfMap = make(map[bson.ObjectId]*torrent.File, len(tracks))
	compare := func(a, b string) int {
		if strings.Contains(b, a) {
			return 0
		}
		return levDistance(a, b)
	}
	for _, track := range tracks {
		var (
			bestMatch    torrent.File
			bestDistance = math.MaxInt32
			bestInd      int
		)
		for ind, file := range files {
			dist := compare(
				strings.ToLower(track.Name),
				strings.ToLower(filepath.Base(file.DisplayPath())),
			)
			if dist < bestDistance {
				bestInd = ind
				bestMatch = file
				bestDistance = dist
			}
		}
		files = append(files[:bestInd], files[bestInd+1:]...)
		tfMap[track.ID] = &bestMatch
	}
	return tfMap
}
