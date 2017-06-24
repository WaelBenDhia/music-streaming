package musicinfo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/wael/music-streaming/models"
)

const root = "http://ws.audioscrobbler.com/2.0"

type lastFMError struct {
	Code    int    `json:"error"`
	Message string `json:"message"`
}
type album struct {
	Name   string `json:"name"`
	Artist string `json:"artist"`
	Image  []struct {
		Text string `json:"#text"`
		Size string `json:"size"`
	} `json:"image"`
	Tracks struct {
		Tracks []track `json:"track"`
	} `json:"tracks"`
}
type track struct {
	Name     string `json:"name"`
	Duration string `json:"duration"`
}
type lastFMAlbumResponse struct {
	Album album `json:"album"`
}

//LastFmClient allows you to make queries to lastFM api
type LastFmClient string

func readBody(r *http.Response) ([]byte, error) {
	return ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
}

//CreateLastFmClient from lastFM api key
func CreateLastFmClient(key string) (LastFmClient, error) {
	resp, err := http.Get(fmt.Sprintf(root+"/?method=album.getinfo&api_key=%s&artist=Cher&album=Believe&format=json", key))
	if err != nil {
		return "", err
	}
	body, err := readBody(resp)
	if err != nil {
		return "", err
	}
	var errResp lastFMError
	if err := json.Unmarshal(body, &errResp); err != nil {
		return "", err
	}
	if errResp.Message != "" {
		return "", errors.New(errResp.Message)
	}
	return LastFmClient(key), err
}

//GetAlbumInfo gets best match for artistName-albumName
func (cli LastFmClient) GetAlbumInfo(artistName, albumName string) (*models.Release, error) {
	query := root + fmt.Sprintf("/?method=album.getinfo&api_key=%s&artist=%s&album=%s&format=json", cli, artistName, albumName)
	resp, err := http.Get(query)
	if err != nil {
		return nil, err
	}
	body, err := readBody(resp)
	if err != nil {
		return nil, err
	}
	var lfmalbum lastFMAlbumResponse
	err = json.Unmarshal(body, &lfmalbum)
	album := parseAlbumFromLFM(lfmalbum)
	return &album, err
}

func parseAlbumFromLFM(resp lastFMAlbumResponse) models.Release {
	var album models.Release
	album.AlbumArtist = &models.Artist{Name: resp.Album.Artist}
	album.Name = resp.Album.Name
	for _, image := range resp.Album.Image {
		if image.Size == "extralarge" {
			album.CoverURL = image.Text
			break
		}
	}
	for _, track := range resp.Album.Tracks.Tracks {
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
