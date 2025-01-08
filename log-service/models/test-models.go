package data

import (
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type MongoTestRepository struct {
	Conn *mongo.Client
}

func NewMongoTestRepository(client *mongo.Client) *MongoTestRepository {
	return &MongoTestRepository{
		Conn: client,
	}
}

// All returns a slice of all logs
func (u *MongoTestRepository) All() ([]*LogEntry, error) {
	logs := []*LogEntry{}

	return logs, nil
}

// GetOne returns one user by id
func (u *MongoTestRepository) GetOne(id string) (*LogEntry, error) {
	log := LogEntry{
		ID:        "1",
		Name:      "Test_Name",
		Data:      "Test_Data",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return &log, nil
}

// Insert inserts a new user into the database, and returns the ID of the newly inserted row
func (u *MongoTestRepository) Insert(log LogEntry) error {
	return nil
}

func (u *MongoTestRepository) UpdateOne(logs LogEntry) (*mongo.UpdateResult, error) {
	return &mongo.UpdateResult{}, nil
}

// DeleteByID deletes one user from the database, by ID
func (u *MongoTestRepository) DropCollection() error {
	return nil
}
