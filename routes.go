package main

func (s *Server) routes() {
	s.router.HandleFunc("/", s.authenticatedOnly(s.handleHomePage))
	s.router.HandleFunc("/favicon.ico", s.handleFavicon)
	s.router.HandleFunc("/login", s.loginRouter)
	s.router.HandleFunc("/logout", s.authenticatedOnly(s.handleLogout))
	s.router.HandleFunc("/register", s.registerRouter)
	s.router.HandleFunc("/upload", s.handleUpload)
	s.router.HandleFunc("/download", s.handleDownload)
}
