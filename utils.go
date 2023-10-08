package main

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

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
