package server

import (
	"io"
	"net/http"
)

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	io.LimitReader()
}
