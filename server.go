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

func newServer(db *gorm.DB, store *sessions.MemoryStore, root string) *Server {
	return &Server{
		db:     db,
		store:  store,
		router: http.NewServeMux(),
		root:   root,
	}
}

type Server struct {
	db     *gorm.DB
	store  *sessions.MemoryStore
	router *http.ServeMux
	root   string // root is the folder name for storing all files
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleFavicon(_ http.ResponseWriter, _ *http.Request) {}

func (s *Server) handleHomePage(w http.ResponseWriter, _ *http.Request,
	_ *sessions.Session) {
	if _, err := os.Stat(s.root); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(s.root, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	type File struct {
		Name string
		Size string
	}
	// get files in the specified root folder
	files, err := os.ReadDir(s.root)
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
	//pass files' info to the HTML template
	tmpl, err := template.ParseFiles("./template/home.html")
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("failed to execute template:%v", err)
	}
}

// Acts as a filter which only allow the logged in request to pass.
func (s *Server) authenticatedOnly(f func(http.ResponseWriter, *http.Request, *sessions.Session)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user has logged in.
		session, err := s.store.Get(r, "session_id")
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to valid request:%v", err), http.StatusInternalServerError)
			log.Printf("Failed to valid request: %v", err)
			return
		}
		isAuth := session.GetValueByKey("authenticated")
		if isAuth == nil {
			http.Error(w, "Not authenticated, session no authenticated field.", http.StatusInternalServerError)
			return
		}
		// Have not logged in, redirect to login page.
		if session.IsNew() || isAuth.(bool) == false {

			return
		}
		// Call the handler.
		f(w, r, session)
	}
}

// Only allow POST method and has password and username parameter to pass.
func (s *Server) postUsernamePasswordOnly(f func(w http.ResponseWriter, r *http.Request,
	username, password string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}
		// Parse username and password in form.
		username, password, err := getUsernameAndPassword(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		f(w, r, username, password)
	}
}

func getUsernameAndPassword(r *http.Request) (username string, password string, err error) {
	// Parse form.
	if e := r.ParseForm(); err != nil {
		err = fmt.Errorf("failed to parse username and password: %v", e)
		return
	}
	// Get username and password form the parsed form.
	username = r.Form.Get("username")
	password = r.Form.Get("password")
	if username == "" || password == "" {
		err = errors.New("no username and password found")
		return
	}
	return
}

// findUser returns an empty User{} if user not found.
func (s *Server) findUser(username string) (*User, error) {
	user := User{}
	// s.db.Limit(1).Find(): returns an empty User{} if user not found.
	err := s.db.Limit(1).Find(&user, "username = ?", username).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}
	if user == (User{}) {
		return nil, errors.New("failed to find user: user doesn't exist")
	}
	return &user, nil
}
