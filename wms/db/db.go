package db

import mgo "gopkg.in/mgo.v2"

//OpenDB opens database on host, returns error if fails
func OpenDB(host, database string) (*mgo.Database, error) {
	session, err := mgo.Dial(host)
	if err != nil {
		return nil, err
	}
	return session.DB(database), nil
}
