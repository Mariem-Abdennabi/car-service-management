// Package assets bridges the Vite build and the Go templates.
//
// Vite writes bundles with a content hash in the filename —
// `assets/app-ZApCAbCH.js` — so that browsers can cache them forever and still
// pick up a new version immediately after a deploy. The trade-off is that no Go
// code can know the filename in advance.
//
// Vite therefore also writes a manifest mapping logical names to real ones. This
// package reads it once at startup and hands the URLs to the templates.
package assets

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

// entryName is the key Vite uses for our bundle: the entry point's path relative
// to the Vite project root (web/).
const entryName = "src/app.js"

// Assets holds the URLs of the built bundle, ready to drop into a template.
type Assets struct {
	JS  string
	CSS string
}

// manifestEntry is the part of Vite's manifest.json that we need. Unlisted fields
// are ignored by encoding/json.
type manifestEntry struct {
	File string   `json:"file"`
	CSS  []string `json:"css"`
}

// Load reads Vite's manifest and returns the URLs for the entry bundle.
//
// urlPrefix is where the server exposes the build directory, e.g. "/build".
//
// It fails loudly. A missing or malformed manifest means the assets were never
// built, and a server that starts anyway would serve every page without styles or
// JavaScript — a confusing failure that looks like a CSS bug. Failing at startup
// points straight at the real problem.
func Load(manifestPath, urlPrefix string) (Assets, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return Assets{}, fmt.Errorf("read asset manifest (run `make assets`): %w", err)
	}

	var manifest map[string]manifestEntry
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Assets{}, fmt.Errorf("parse asset manifest %s: %w", manifestPath, err)
	}

	entry, ok := manifest[entryName]
	if !ok {
		return Assets{}, fmt.Errorf("asset manifest %s has no entry for %q", manifestPath, entryName)
	}
	if len(entry.CSS) == 0 {
		return Assets{}, fmt.Errorf("asset manifest entry %q lists no stylesheet", entryName)
	}

	return Assets{
		JS:  path.Join(urlPrefix, entry.File),
		CSS: path.Join(urlPrefix, entry.CSS[0]),
	}, nil
}
