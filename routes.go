package main

func (s *Server) routes() {
	s.router.HandleFunc("upload", s.handleUpload)
	s.router.HandleFunc("download", s.handleDownload)
	s.router.HandleFunc("/favicon.ico", s.handleFavicon)
	s.router.HandleFunc("/", s.authenticatedOnly(s.handleHomePage))
	s.router.HandleFunc("logout", s.authenticatedOnly(s.handleAuthLogout))
	s.router.HandleFunc("/login", s.postUsernamePasswordOnly(s.handleAuthLogin))
	s.router.HandleFunc("/register", s.postUsernamePasswordOnly(s.handleRegister))
}
