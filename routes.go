package main

import (
	"fmt"
	"net/http"
)

func test(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hello")
}

func (s *Server) routes() {
	s.router.HandleFunc("/", s.handleHomePage)
	s.router.HandleFunc("/favicon.ico", s.handleFavicon)
	s.router.HandleFunc("/login", s.loginRouter)
	s.router.HandleFunc("/logout", s.authenticatedOnly(s.handleLogout))
	s.router.HandleFunc("/register", s.registerRouter)
	s.router.HandleFunc("/upload", s.authenticatedOnly(s.handleUpload))
	s.router.HandleFunc("/download", s.authenticatedOnly(s.handleDownload))
}
