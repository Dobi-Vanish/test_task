package data

import "go.mongodb.org/mongo-driver/mongo"

type Repository interface {
	Insert(entry LogEntry) error
	All() ([]*LogEntry, error)
	GetOne(id string) (*LogEntry, error)
	DropCollection() error
	UpdateOne(logs LogEntry) (*mongo.UpdateResult, error)
}
