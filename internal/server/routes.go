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
	s.router.GET("/customers", s.handleCustomers)
	s.router.GET("/customers/new", s.handleNewCustomer)
	s.router.POST("/customers", s.handleCreateCustomer)
	s.router.GET("/customers/:id", s.handleCustomer)
	s.router.GET("/customers/:id/edit", s.handleEditCustomer)
	s.router.POST("/customers/:id", s.handleUpdateCustomer)
	s.router.POST("/customers/:id/delete", s.handleDeleteCustomer)
	s.router.GET("/customers/:id/vehicles/new", s.handleNewVehicle)
	s.router.POST("/customers/:id/vehicles", s.handleCreateVehicle)

	// Not nested: a vehicle id is already unique, and /customers/1/vehicles/2 would
	// allow a URL where vehicle 2 belongs to someone else. The owner is read from
	// the vehicle instead.
	s.router.GET("/vehicles/:id/edit", s.handleEditVehicle)
	s.router.POST("/vehicles/:id", s.handleUpdateVehicle)
	s.router.POST("/vehicles/:id/delete", s.handleDeleteVehicle)

	// A job is read on its own unnested URL for the same reason a vehicle is: the
	// id is unique, and nesting it under a customer would allow a URL where the two
	// disagree. Opening one is nested under its vehicle — step 6b.
	s.router.GET("/jobs", s.handleJobs)
	s.router.GET("/jobs/:id", s.handleJob)

	// Operational
	s.router.GET("/parts", s.handleParts)
	s.router.GET("/parts/new", s.handleNewPart)
	s.router.POST("/parts", s.handleCreatePart)
	s.router.GET("/parts/:id/edit", s.handleEditPart)
	s.router.POST("/parts/:id", s.handleUpdatePart)
	s.router.POST("/parts/:id/delete", s.handleDeletePart)

	s.router.GET("/healthz", s.handleHealth)

	// Anything unmatched gets the same 404 page as a missing record.
	s.router.NoRoute(s.notFound)
}
