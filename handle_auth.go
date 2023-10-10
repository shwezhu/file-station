package main

import (
	"fmt"
	"github.com/shwezhu/sessions"
	"log"
	"net/http"
)

func (s *Server) loginRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, fmt.Sprintf("%v/login.html", s.templateRoot))
	} else {
		s.postUsernamePasswordOnly(s.handleLoginPost)(w, r)
	}
}

func (s *Server) registerRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, fmt.Sprintf("%v/register.html", s.templateRoot))
	} else {
		s.postUsernamePasswordOnly(s.handleRegister)(w, r)
	}
}

func (s *Server) handleLoginPost(w http.ResponseWriter, r *http.Request,
	username, password string) {
	// Query user by username in database.
	user, err := s.findUser(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	// No such user.
	if *user == (User{}) {
		http.Error(w, "You have not registered yet.", http.StatusUnauthorized)
		return
	}
	// Compare provided password with password stored in database.
	if !comparePasswordHash(user.Password, password) {
		http.Error(w, "Password is incorrect.", http.StatusUnauthorized)
		return
	}

	session, err := s.store.Get(r, "session_id")
	// User has logged in.
	if !session.IsNew() {
		_, _ = fmt.Fprint(w, "")
		return
	}

	// User hasn't logged in, config session.
	session.SetMaxAge(30 * 60)
	session.SetValue("username", username)
	// Save session.
	session.Save(w)
	// Login successfully.
	_, _ = fmt.Fprint(w, "")
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request,
	username, password string) {
	http.Error(w, "Register is not allowed temporarily", http.StatusUnauthorized)
	return
	/*	// Store encrypted password in database.
		hashedPassword, err := hashPassword(password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Check if the user table exists, if not, create one.
		err = s.validUserTable()
		if err != nil {
			log.Fatal("Failed to migrate the user schema.")
		}
		// Check if user has existed in the database.
		user, err := s.findUser(username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		if *user != (User{}) {
			http.Error(w, "Username has been taken.", http.StatusConflict)
			return
		}
		user.Username = username
		user.Password = hashedPassword
		// Save user into database.
		if err = s.db.Create(user).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
		}
		// Register successfully.
		_, _ = fmt.Fprint(w, "")*/
}

func (s *Server) handleLogout(w http.ResponseWriter, _ *http.Request,
	session *sessions.Session) {
	session.SetMaxAge(-1)
	session.SetIsNew(true)
	session.Save(w)
	// Logout successfully.
	_, _ = fmt.Fprint(w, "")
}
