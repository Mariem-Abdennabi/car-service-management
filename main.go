// Command car-service-management runs the Car Service & Spare Parts Management
// System.
//
// This file is the composition root: it reads configuration, builds the pieces,
// wires them together, and starts the server. Keeping it that way means one file
// answers "how is this application assembled?"
package main

import (
	"context"
	"embed"
	"flag"
	"log"

	"github.com/Mariem-Abdennabi/car-service-management/assets"
	"github.com/Mariem-Abdennabi/car-service-management/internal/config"
	"github.com/Mariem-Abdennabi/car-service-management/internal/server"
	"github.com/Mariem-Abdennabi/car-service-management/internal/store"
)

// Where Vite writes its manifest, and the URL path the build directory is served
// at.
const (
	manifestPath = "public/build/.vite/manifest.json"
	assetsPrefix = "/build"
)

// The schema, compiled into the binary. A deployed binary therefore carries its
// own migrations: one artefact to ship, and no way to run new code against an old
// database.
//
//go:embed sql/migrations/*.sql
var migrations embed.FS

func main() {
	seed := flag.Bool("seed", false, "replace all data with demo data, then exit")
	flag.Parse()

	if err := config.LoadFile(".env"); err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	// Before the pool, so the schema is correct before anything queries it.
	if err := store.Migrate(migrations, cfg.DatabaseURL); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	db, err := store.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	// Runs when main returns, which for a clean exit means never — the process is
	// killed while ListenAndServe blocks. It matters for the paths that do return:
	// a server that fails to bind its port closes the pool on the way out.
	defer db.Close()

	if *seed {
		// Seeding deletes everything. The environment check is the only thing
		// standing between a stray flag and a wiped production database.
		if !cfg.IsDevelopment() {
			log.Fatalf("refusing to seed in %s mode", cfg.Env)
		}

		if err := db.Seed(context.Background()); err != nil {
			log.Fatalf("seed error: %v", err)
		}

		log.Println("seeded with demo data")

		return
	}

	// After the seed branch, so seeding does not need the frontend built.
	builtAssets, err := assets.Load(manifestPath, assetsPrefix)
	if err != nil {
		log.Fatalf("assets error: %v", err)
	}

	log.Printf("car-service-management: starting in %s mode", cfg.Env)

	if err := server.New(cfg, builtAssets, db).Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
