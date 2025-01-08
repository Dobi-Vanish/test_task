package main

import (
	"log-service/models"
	"os"
	"testing"
)

var testApp Config

func TestMain(m *testing.M) {
	repo := data.NewMongoTestRepository(nil)
	testApp.Repo = repo
	os.Exit(m.Run())
}
