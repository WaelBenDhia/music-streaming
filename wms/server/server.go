package server

import (
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/wael/music-streaming/lastfm"
	"github.com/wael/music-streaming/wms/models"
	"github.com/wael/music-streaming/wms/torrent"
)

type middleware func(http.Handler) http.Handler

//Server is a music-streaming server
type Server struct {
	http.Handler
	server                        *http.Server
	infoLog, warningLog, errorLog *log.Logger
	db                            *sql.DB
	lfmCli                        *lastfm.Client
	torrentCli                    *torrent.Client
}

//NewServer creates and initializes a new music streaming server
func NewServer(stdOut, stdErr io.Writer, dbPath, lastFMApiKey, downDir, listenAddr string) (Server, error) {
	s := Server{}
	return s, s.init(stdOut, stdErr, dbPath, lastFMApiKey, downDir, listenAddr)
}

//Start server
func (s *Server) Start(listenAddr string) <-chan int {
	s.server = &http.Server{Addr: listenAddr, Handler: s}
	doneChan := make(chan int, 1)
	go func() {
		s.infoLog.Println("Starting server, listening on address:", s.server.Addr)
		var exitVal int
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			s.errorLog.Printf("Could not start server: %v", err)
			exitVal = 1
		}
		doneChan <- exitVal
		close(doneChan)
	}()
	return doneChan
}

//Stop the server
func (s *Server) Stop() error {
	if err := s.server.Shutdown(nil); err != nil {
		return err
	}
	s.server = nil
	return s.closeDB()
}

func (s *Server) init(stdOut, stdErr io.Writer, dbPath, lastFMApiKey, downDir, listenAddr string) error {
	s.initLogging(stdOut, stdErr)
	s.initRouting()
	err := s.initDB(dbPath)
	if err != nil {
		return err
	}
	err = s.initlfmCli(lastFMApiKey)
	if err != nil {
		return err
	}
	return s.initTorrentClient(downDir, listenAddr)
}

func (s *Server) initLogging(stdOut, stdErr io.Writer) {
	s.infoLog = log.New(stdOut, "INFO:", log.Ldate|log.Ltime)
	s.warningLog = log.New(stdOut, "WARNING:", log.Ldate|log.Ltime)
	s.errorLog = log.New(stdErr, "ERROR:", log.Ldate|log.Ltime)
}

func (s *Server) initRouting() {
	router := mux.NewRouter().StrictSlash(true)
	s.infoLog.Println("Registering endpoints.")
	for _, endpoint := range []struct {
		method, name, path string
		handler            http.Handler
		middlewares        []middleware
	}{
		{
			"GET",
			"NOTHING",
			"/",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("hello"))
			}),
			nil,
		},
	} {
		s.infoLog.Printf("Registering '%s' endpoint: '%s': %s", endpoint.name, endpoint.path, endpoint.path)
		router.
			Methods(endpoint.method).
			Path(endpoint.path).
			Name(endpoint.name).
			Handler(AddMiddleware(endpoint.handler)(endpoint.middlewares...))
	}
	s.Handler = router
}

func (s *Server) initDB(DBPath string) error {
	if s.db != nil {
		return nil
	}
	s.infoLog.Println("Initializing Database")
	if _, err := os.Stat(DBPath); err != nil {
		s.warningLog.Println("Could not find Database:'" + DBPath + "'. Creating it.")
		if _, err := os.Create(DBPath); err != nil {
			return err
		}
	}
	var err error
	s.db, err = sql.Open("sqlite3", DBPath)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	s.infoLog.Println("Registering entities.")
	for _, entity := range sortEntitiesByPriority(
		&models.Artist{},
		&models.Release{},
		&models.Statistic{},
		&models.Track{}) {
		if err == nil {
			err = entity.CreateTable(tx)
		}
	}
	if err == nil {
		return tx.Commit()
	}
	_ = tx.Rollback() //We are ignoring the error here because we care about the already existing err
	s.infoLog.Println("Initialized database successfully.")
	return err
}

func (s *Server) closeDB() error {
	if s.db == nil {
		s.warningLog.Println("Tried to close already closed database")
		return nil
	}
	return s.db.Close()
}

func (s *Server) initlfmCli(apiKey string) error {
	cli, err := lastfm.CreateLastFmClient(apiKey)
	if err != nil {
		s.errorLog.Printf("Could not create last FM Client: %v", err)
	}
	s.lfmCli = &cli
	return err
}

func (s *Server) initTorrentClient(downloadDirectory, listenAddr string) error {
	cli, err := torrent.NewClient(downloadDirectory, listenAddr)
	if err != nil {
		s.errorLog.Printf("Could not create torrent Client: %v", err)
	}
	s.torrentCli = &cli
	return err
}
