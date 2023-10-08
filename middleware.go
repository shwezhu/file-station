package main

import (
	"fmt"
	"github.com/shwezhu/sessions"
	"log"
	"net/http"
)

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
		// Have not logged in, redirect to login page.
		if session.IsNew() {
			http.Redirect(w, r, "http://localhost:8080/login", http.StatusPermanentRedirect)
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
