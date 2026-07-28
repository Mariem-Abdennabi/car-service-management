// Package server owns the HTTP layer: the router, the middleware, and the
// dependencies handlers need.
package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Mariem-Abdennabi/car-service-management/assets"
	"github.com/Mariem-Abdennabi/car-service-management/internal/config"
)

// Server holds everything the HTTP handlers need.
//
// Handlers are methods on this struct, so they reach configuration, the built
// asset URLs — and, from Milestone 2, the database pool — through the receiver
// rather than through package-level variables. That is what keeps them testable: a
// test constructs its own Server instead of arranging global state.
type Server struct {
	cfg    config.Config
	assets assets.Assets
	router *gin.Engine
}

// New builds a Server with its middleware and routes already registered.
func New(cfg config.Config, builtAssets assets.Assets) *Server {
	if !cfg.IsDevelopment() {
		// Suppresses Gin's start-up banner and debug warnings.
		gin.SetMode(gin.ReleaseMode)
	}

	// gin.New gives a router with no middleware, unlike gin.Default, so what is
	// installed here is exactly what runs.
	router := gin.New()

	// Recovery turns a panic in a handler into a 500 for that one request
	// instead of a crashed process. Always on.
	router.Use(gin.Recovery())

	// A line per request is useful while building and noise in production, where
	// a reverse proxy usually logs the same thing.
	if cfg.IsDevelopment() {
		router.Use(gin.Logger())
	}

	s := &Server{cfg: cfg, assets: builtAssets, router: router}
	s.registerRoutes()

	return s
}

// Run starts the HTTP server and blocks until it stops.
//
// Ctrl-C ends the process immediately, dropping any request in flight. That is
// fine while developing; making shutdown graceful is a Milestone 9 concern and
// is recorded in docs/backlog.md.
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: s.router,
		// Without this, a client can hold a connection open indefinitely by
		// sending headers one byte at a time.
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("listening on http://localhost%s", addr)

	return httpServer.ListenAndServe()
}
