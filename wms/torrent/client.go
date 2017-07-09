package torrent

import (
	"errors"
	"log"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/wael/music-streaming/gopirate"
)

//ErrTorrentNotFound if torrent is not found this error is returned
var ErrTorrentNotFound = errors.New("torrent not found")

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
		Seed:       true,
		Debug:      true,
	})
	return Client{cli, make(map[string]*torrent.Torrent)}, err
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

//GetInfo returns torrent metadata if exists
func (cli *Client) GetInfo(torrent gopirate.Torrent) *metainfo.Info {
	if tor := cli.getTorrent(torrent); tor != nil {
		return tor.Info()
	}
	return nil
}

//StartAll downloads all files within given torrent
func (cli *Client) StartAll(torrent gopirate.Torrent) error {
	if tor := cli.getTorrent(torrent); tor != nil {
		tor.DownloadAll()
		return nil
	}
	return ErrTorrentNotFound
}

//PrintStatus prints status of given torrent
func (cli *Client) PrintStatus(torrent gopirate.Torrent, logger *log.Logger) {
	logger.Println(cli.getTorrent(torrent).Stats())
	logger.Println("Remaining ", cli.getTorrent(torrent).BytesMissing())
}

//IsComplete returns true if torrent has finished downloading
func (cli *Client) IsComplete(torrent gopirate.Torrent) bool {
	if tor := cli.getTorrent(torrent); tor != nil {
		return tor.BytesMissing() > 0
	}
	return false
}
