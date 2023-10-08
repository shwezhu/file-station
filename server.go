package main

import (
	"errors"
	"fmt"
	"github.com/shwezhu/sessions"
	"gorm.io/gorm"
	"html/template"
	"log"
	"net/http"
	"os"
)

func newServer(db *gorm.DB, store *sessions.MemoryStore) *Server {
	return &Server{
		db:     db,
		store:  store,
		router: http.NewServeMux(),
	}
}

type Server struct {
	db           *gorm.DB
	store        *sessions.MemoryStore
	router       *http.ServeMux
	fileRoot     string // fileRoot is the folder name for storing all files
	templateRoot string
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleFavicon(_ http.ResponseWriter, _ *http.Request) {}

func (s *Server) handleHomePage(w http.ResponseWriter, _ *http.Request,
	_ *sessions.Session) {
	if _, err := os.Stat(s.fileRoot); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(s.fileRoot, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	type File struct {
		Name string
		Size string
	}
	// get files in the specified root folder
	files, err := os.ReadDir(s.fileRoot)
	if err != nil {
		http.Error(w, fmt.Errorf("failed to open root folder: %v", err).Error(), http.StatusInternalServerError)
		return
	}
	data := make([]File, 0)
	for _, f := range files {
		// ignore folder
		if f.IsDir() {
			continue
		}
		info, err := f.Info()
		if err != nil {
			log.Printf("failed to get info of %v: %v", f.Name(), err)
			continue
		}
		// size less than 100KB, displayed as KB
		var size string
		if info.Size() <= 102400 {
			size = fmt.Sprintf("%.2f KB", float32(info.Size())/(1<<10))
		} else {
			size = fmt.Sprintf("%.2f MB", float32(info.Size())/(1<<20))
		}
		data = append(data, File{
			Name: f.Name(),
			Size: size,
		})
	}
	// Pass files' info to the HTML template
	tmpl, err := template.ParseFiles("./template/home.html")
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("failed to execute template:%v", err)
	}
}
