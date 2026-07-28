package assets

import (
	"os"
	"path/filepath"
	"testing"
)

// writeManifest puts contents into a manifest file inside a temporary directory
// and returns its path.
func writeManifest(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write test manifest: %v", err)
	}
	return path
}

func TestLoad(t *testing.T) {
	const valid = `{
		"src/app.js": {
			"file": "assets/app-ABC123.js",
			"isEntry": true,
			"css": ["assets/app-DEF456.css"]
		}
	}`

	t.Run("resolves hashed filenames to URLs", func(t *testing.T) {
		got, err := Load(writeManifest(t, valid), "/build")
		if err != nil {
			t.Fatalf("Load() returned an unexpected error: %v", err)
		}

		want := Assets{
			JS:  "/build/assets/app-ABC123.js",
			CSS: "/build/assets/app-DEF456.css",
		}
		if got != want {
			t.Errorf("Load() = %+v, want %+v", got, want)
		}
	})

	// Each of these would otherwise produce a server that starts happily and
	// serves every page unstyled.
	t.Run("rejects a missing file", func(t *testing.T) {
		if _, err := Load(filepath.Join(t.TempDir(), "absent.json"), "/build"); err == nil {
			t.Error("Load() succeeded on a missing manifest, want an error")
		}
	})

	t.Run("rejects malformed JSON", func(t *testing.T) {
		if _, err := Load(writeManifest(t, "not json"), "/build"); err == nil {
			t.Error("Load() succeeded on malformed JSON, want an error")
		}
	})

	t.Run("rejects a manifest without our entry point", func(t *testing.T) {
		if _, err := Load(writeManifest(t, `{"src/other.js": {"file": "x.js"}}`), "/build"); err == nil {
			t.Error("Load() succeeded without the entry point, want an error")
		}
	})

	t.Run("rejects an entry with no stylesheet", func(t *testing.T) {
		if _, err := Load(writeManifest(t, `{"src/app.js": {"file": "app.js"}}`), "/build"); err == nil {
			t.Error("Load() succeeded with no stylesheet, want an error")
		}
	})
}
