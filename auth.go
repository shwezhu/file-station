package main

import (
	"fmt"
	"github.com/shwezhu/sessions"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

// If user has logged in, redirect to home page.
func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request,
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
	// User has logged in, redirect to home page.
	if !session.IsNew() {

		return
	}

	// User hasn't logged in, config session.
	session.SetMaxAge(30 * 60)
	session.SetValue("authenticated", true)
	session.SetValue("username", username)
	// Save session.
	session.Save(w)
	// Login successfully, redirect to home page.
	if _, err = fmt.Fprint(w, "login successfully"); err != nil {
		log.Println(err)
		return
	}
}

func (s *Server) handleAuthLogout(w http.ResponseWriter, _ *http.Request,
	session *sessions.Session) {
	session.SetMaxAge(-1)
	session.SetValue("authenticated", false)
	session.Save(w)
	// Redirect to login page.
	if _, err := fmt.Fprint(w, "logout successfully"); err != nil {
		log.Println(err)
		return
	}
}

// Redirect to login page
func (s *Server) handleRegister(w http.ResponseWriter, _ *http.Request,
	username, password string) {
	// Store encrypted password in database.
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
	// Redirect to login page.
	if _, err = fmt.Fprint(w, "registered successfully"); err != nil {
		log.Println(err)
		return
	}
}

// validUserTable checks if user table exists, if not, create one.
func (s *Server) validUserTable() error {
	if !s.db.Migrator().HasTable(&User{}) {
		// Migrate the schema - create table.
		return s.db.AutoMigrate(&User{})
	}
	return nil
}

func comparePasswordHash(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}
