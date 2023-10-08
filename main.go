package main

import (
	"github.com/shwezhu/sessions"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func main() {
	db, err := gorm.Open(sqlite.Open("file_station.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}
	store := sessions.NewMemoryStore(sessions.WithDefaultGc())
	if err != nil {
		log.Fatal("failed to create Redis store")
	}
	srv := newServer(db, store)
	srv.fileRoot = "./files"
	srv.templateRoot = "./template"
	_ = http.ListenAndServe(":8080", srv)
}
