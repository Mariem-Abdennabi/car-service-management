package server

// registerRoutes lists every URL the application answers.
//
// Keeping them in one function means there is a single place to answer "what
// does this application respond to?" — a question that gets harder to answer
// every time routes are scattered across the files that implement them.
func (s *Server) registerRoutes() {
	// The Vite bundle. The path is relative to the process's working directory, so
	// the server must be started from the repository root — which `make run` does.
	s.router.Static("/build", "public/build")

	// Pages
	s.router.GET("/", s.handleHome)

	// Operational
	s.router.GET("/healthz", s.handleHealth)
}
