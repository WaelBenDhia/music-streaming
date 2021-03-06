package server

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

type key int

const (
	requestKey key = iota
)

func (s *Server) requestParsingMiddleware(v interface{}) middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
			if err != nil {
				http.Error(w, "Error parsing request body", 400)
				s.errorLog.Printf(
					"requestParsingMiddleware: error reading from body: %v",
					err,
				)
				return
			}
			if err := r.Body.Close(); err != nil {
				w.WriteHeader(500)
				w.Header().Set("Content-Type", "text/plain")
				s.errorLog.Printf(
					"requestParsingMiddleware: error in closing body: %v",
					err,
				)
				return
			}
			if err := json.Unmarshal(body, v); err != nil {
				w.WriteHeader(400)
				s.errorLog.Printf(
					"requestParsingMiddleware: error unmarshaling from body: %v", 
					err,
				)
				return
			}

			ctx, ctxCancel := ctxWithValCancel(r.Context(), requestKey, v)
			defer ctxCancel()
			// Rewrite body in case it's needed down the line
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
