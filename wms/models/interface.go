package models

import (
	"gopkg.in/mgo.v2"
)

//ColCreator is a struct that has a method that creates a collection
type ColCreator interface {
	ColCreate(*mgo.Database) error
}

//Savable to database
type Savable interface {
	Save(*mgo.Database) error
}
