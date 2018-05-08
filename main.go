package main

import (
	"log"
	"os"

	"github.com/waelbendhia/music-streaming/wms/server"
)

func main() {
	server, err := server.NewServer(
		os.Stdout,
		os.Stderr,
		"localhost",
		"wmsDB",
		os.Getenv("LASTFM_API_KEY"),
		"/home/wael/third-world-streams/",
		"0.0.0.0:12345",
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server exited with value:", <-server.Start(":8082"))
}
