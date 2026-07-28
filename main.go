// Command car-service-management runs the Car Service & Spare Parts Management
// System.
//
// This file is the composition root: it reads configuration, builds the pieces,
// wires them together, and starts the server. Keeping it that way means one file
// answers "how is this application assembled?"
package main

import (
	"log"

	"github.com/Mariem-Abdennabi/car-service-management/assets"
	"github.com/Mariem-Abdennabi/car-service-management/internal/config"
	"github.com/Mariem-Abdennabi/car-service-management/internal/server"
)

// Where Vite writes its manifest, and the URL path the build directory is served
// at.
const (
	manifestPath = "public/build/.vite/manifest.json"
	assetsPrefix = "/build"
)

func main() {
	if err := config.LoadFile(".env"); err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	builtAssets, err := assets.Load(manifestPath, assetsPrefix)
	if err != nil {
		log.Fatalf("assets error: %v", err)
	}

	log.Printf("car-service-management: starting in %s mode", cfg.Env)

	if err := server.New(cfg, builtAssets).Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
