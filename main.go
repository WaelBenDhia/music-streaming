package main

import (
	"log"
	"os"
	"time"

	"github.com/wael/music-streaming/musicinfo"
	"github.com/wael/music-streaming/torrentclient"
	"github.com/wael/music-streaming/tpbclient"
)

func main() {
	log.Print("HELLO")
	// log.Print(musicinfo.CreateLastFmClient("DICKS"))
	log.Print(os.Getenv("LASTFM_API_KEY"))
	cli, _ := musicinfo.CreateLastFmClient(os.Getenv("LASTFM_API_KEY"))
	album, _ := cli.GetAlbumInfo("Mobb Deep", "The Infamous")
	results, err := tpbclient.Search(album.AlbumArtist.Name + " " + album.Name)
	if err != nil {
		panic(err)
	}
	if len(results) == 0 {
		panic("No results")
	} else {
		log.Println("RESULTS::::", len(results))
		for _, result := range results {
			log.Println(result)
		}
	}
	torCli, err := torrentclient.NewClient("./", "localhost:12345")
	if err != nil {
		panic(err)
	}
	tor, err := torCli.AddMagnet(results[0].Link)
	if err != nil {
		panic(err)
	}
	ok := true
	for ok {
		_, ok = <-tor.GotInfo()
	}
	for _, file := range tor.Info().Files {
		log.Println(file)
	}
	tor.DownloadAll()
	for {
		log.Println(torCli.Torrents()[0].BytesCompleted(), "::", torCli.Torrents()[0].Length())
		time.Sleep(time.Second)
	}
}
