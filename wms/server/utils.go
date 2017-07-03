package server

import (
	"context"
	"net/http"
)

//AddMiddleware creates a new handler adapted with middleware
func AddMiddleware(h http.Handler) func(...middleware) http.Handler {
	return func(ads ...middleware) http.Handler {
		for _, mw := range ads {
			h = mw(h)
		}
		return h
	}
}

func ctxWithValCancel(ctx context.Context, valKey key, val interface{}) (context.Context, context.CancelFunc) {
	return context.WithCancel(context.WithValue(ctx, valKey, val))
}
