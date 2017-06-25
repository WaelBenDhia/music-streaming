package server

import (
	"math/rand"
	"net/http"

	"github.com/wael/music-streaming/models"
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

func sortEntitiesByPriority(entities ...models.Creator) []models.Creator {
	if len(entities) < 2 {
		return entities
	}

	left, right := 0, len(entities)-1

	// Pick a pivot
	pivotIndex := rand.Int() % len(entities)

	// Move the pivot to the right
	entities[pivotIndex], entities[right] = entities[right], entities[pivotIndex]

	// Pile elements smaller than the pivot on the left
	for i := range entities {
		if entities[i].CreatePriority() < entities[right].CreatePriority() {
			entities[i], entities[left] = entities[left], entities[i]
			left++
		}
	}

	// Place the pivot after the last smaller element
	entities[left], entities[right] = entities[right], entities[left]

	// Go down the rabbit hole
	sortEntitiesByPriority(entities[:left]...)
	sortEntitiesByPriority(entities[left+1:]...)
	return entities
}
