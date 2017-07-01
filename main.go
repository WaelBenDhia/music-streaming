package main

import (
	"log"
	"os"

	"github.com/wael/music-streaming/wms/server"
)

func main() {
	server, err := server.NewServer(os.Stdout, os.Stderr, "./streaming.db", os.Getenv("LASTFM_API_KEY"), "./downloads", ":12345")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("server exited with value:", <-server.Start(":8082"))
}
