package lastfm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

//SearchResults of a last FM api search
type SearchResults struct {
	Results struct {
		OpensearchQuery struct {
			Text        string `json:"#text"`
			Role        string `json:"role"`
			SearchTerms string `json:"searchTerms"`
			StartPage   string `json:"startPage"`
		} `json:"opensearch:Query"`
		OpensearchTotalResults string `json:"opensearch:totalResults"`
		OpensearchStartIndex   string `json:"opensearch:startIndex"`
		OpensearchItemsPerPage string `json:"opensearch:itemsPerPage"`
		AlbumMatches           struct {
			Album []Album `json:"album"`
		} `json:"albummatches"`
	} `json:"results"`
}

const root = "http://ws.audioscrobbler.com/2.0"

//Error from last FM api
type Error struct {
	Code    int    `json:"error"`
	Message string `json:"message"`
}

//Album from last FM api
type Album struct {
	Name   string `json:"name"`
	Artist string `json:"artist"`
	URL    string `json:"url"`
	Image  []struct {
		Text string `json:"#text"`
		Size string `json:"size"`
	} `json:"image"`
	Tracks struct {
		Tracks []Track `json:"track"`
	} `json:"tracks"`
	Streamable string `json:"streamable"`
	MBID       string `json:"mbid"`
}

//Track from last FM api
type Track struct {
	Name     string `json:"name"`
	Duration string `json:"duration"`
}

//AlbumResponse from last FM api
type AlbumResponse struct {
	Album Album `json:"album"`
}

//Client allows you to make queries to lastFM api
type Client string

func readBody(r *http.Response) ([]byte, error) {
	return ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
}

//CreateLastFmClient from lastFM api key
func CreateLastFmClient(key string) (Client, error) {
	resp, err := http.Get(fmt.Sprintf(
		root+"/?method=album.getinfo"+
			"&api_key=%s"+
			"&artist=Cher"+
			"&album=Believe"+
			"&format=json",
		key,
	))
	if err != nil {
		return "", err
	}
	body, err := readBody(resp)
	if err != nil {
		return "", err
	}
	var errResp Error
	if err := json.Unmarshal(body, &errResp); err != nil {
		return "", err
	}
	if errResp.Message != "" {
		return "", errors.New(errResp.Message)
	}
	return Client(key), err
}

//GetAlbumInfo gets best match for artistName-albumName
func (cli Client) GetAlbumInfo(artistName, albumName string) (*AlbumResponse, error) {
	query := root + fmt.Sprintf(
		"/?method=album.getinfo&api_key=%s&artist=%s&album=%s&format=json",
		cli,
		artistName,
		albumName,
	)
	resp, err := http.Get(query)
	if err != nil {
		return nil, err
	}
	body, err := readBody(resp)
	if err != nil {
		return nil, err
	}
	var lfmalbum AlbumResponse
	err = json.Unmarshal(body, &lfmalbum)
	return &lfmalbum, err
}

//SearchAlbums searches for albums
func (cli Client) SearchAlbums(searchTerm string) (*SearchResults, error) {
	query := root + fmt.Sprintf(
		"/?method=album.search&api_key=%s&album=%s&format=json",
		cli,
		searchTerm,
	)
	resp, err := http.Get(query)
	if err != nil {
		return nil, err
	}
	body, err := readBody(resp)
	if err != nil {
		return nil, err
	}
	var results SearchResults
	err = json.Unmarshal(body, &results)
	return &results, err
}

// func (lfmAlbum *Album) toWMSAlbum() models.Release {
// 	var album models.Release
// 	album.AlbumArtist = &models.Artist{Name: lfmAlbum.Artist}
// 	album.Name = lfmAlbum.Name
// 	for _, image := range lfmAlbum.Image {
// 		if image.Size == "extralarge" {
// 			album.CoverURL = image.Text
// 			break
// 		}
// 	}
// 	for _, track := range lfmAlbum.Tracks.Tracks {
// 		var newTrack models.Track
// 		duration, convErr := strconv.Atoi(track.Duration)
// 		if convErr != nil {
// 			duration = 0
// 		}
// 		newTrack.Length = time.Duration(duration) * time.Second
// 		newTrack.Name = track.Name
// 		album.Tracks = append(album.Tracks, newTrack)
// 	}
// 	return album
// }
