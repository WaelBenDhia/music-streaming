package torrent

import (
	"errors"

	"github.com/anacrolix/torrent"
	"github.com/wael/music-streaming/gopirate"
	"github.com/wael/music-streaming/wms/models"
)

//Client is a torrent client
type Client struct {
	*torrent.Client
	torrents map[string]*torrent.Torrent
}

//NewClient creates a new torrent client
func NewClient(downloadDirectory, listenAddr string) (Client, error) {
	cli, err := torrent.NewClient(&torrent.Config{
		DataDir:    downloadDirectory,
		ListenAddr: listenAddr,
		NoUpload:   false,
		Seed:       false,
		Debug:      true,
	})
	return Client{cli, make(map[string]*torrent.Torrent)}, err
}

//FindAndAddTPBTorrent search for relase on TPB and adds it to client
func (cli *Client) FindAndAddTPBTorrent(release models.Release) error {
	results, err := searchRelease(release)
	if err != nil {
		return err
	}
	return cli.AddTPBTorrent(results[0])
}

//AddTPBTorrent adds a magnet link to client
func (cli *Client) AddTPBTorrent(torrent gopirate.Torrent) error {
	tor, err := cli.AddMagnet(torrent.Link)
	if err == nil {
		cli.torrents[torrent.Link] = tor
	}
	return err
}

func (cli *Client) getTorrent(torrent gopirate.Torrent) *torrent.Torrent {
	tor, found := cli.torrents[torrent.Link]
	if found {
		return tor
	}
	return nil
}

//GotInfo returns a channel that closes when torrent has info, or nil if torrent has not been added
func (cli *Client) GotInfo(torrent gopirate.Torrent) <-chan struct{} {
	if tor := cli.getTorrent(torrent); tor != nil {
		return tor.GotInfo()
	}
	return nil
}

//IsComplete returns true if torrent has finished downloading
func (cli *Client) IsComplete(torrent gopirate.Torrent) bool {
	if tor := cli.getTorrent(torrent); tor != nil {
		return tor.BytesMissing() > 0
	}
	return false
}

//searchRelease warps gopirate.Search to receive models.release
func searchRelease(release models.Release) ([]gopirate.Torrent, error) {
	if release.AlbumArtist == nil {
		return nil, errors.New("FindAndAddTPBTorrent: release has no artist associated")
	}
	return gopirate.Search(release.AlbumArtist.Name + " " + release.Name)
}
