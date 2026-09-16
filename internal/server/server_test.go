package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Mariem-Abdennabi/car-service-management/assets"
	"github.com/Mariem-Abdennabi/car-service-management/internal/config"
	"github.com/Mariem-Abdennabi/car-service-management/internal/store"
)

// newTestServer builds a Server for tests.
//
// Production mode keeps Gin's per-request log lines out of the test output. The
// port is irrelevant because these tests drive the router directly and never bind
// a socket, and the asset URLs are fixed rather than read from a real Vite
// manifest — the point is that the layout renders whatever it is given.
//
// db may be nil for a page that does not read the database. Most now do.
func newTestServer(db *store.Store) *Server {
	return New(
		config.Config{Env: "production", Port: 8080},
		assets.Assets{JS: "/build/app.js", CSS: "/build/app.css"},
		db,
	)
}

// do sends a request through the router without starting a real HTTP server.
// httptest.NewRecorder stands in for the response writer, so these tests need no
// network at all and run in microseconds.
func do(t *testing.T, db *store.Store, method, path string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	newTestServer(db).router.ServeHTTP(recorder, httptest.NewRequest(method, path, nil))

	return recorder
}

// TestHandleHealth needs a real database, since the whole point of the endpoint
// now is that it reports the database's reachability.
func TestHandleHealth(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping database tests")
	}

	db, err := store.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	got := do(t, db, http.MethodGet, "/healthz")

	if got.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", got.Code, http.StatusOK)
	}
	if want := `{"status":"ok"}`; got.Body.String() != want {
		t.Errorf("body = %s, want %s", got.Body.String(), want)
	}
	if want := "application/json; charset=utf-8"; got.Header().Get("Content-Type") != want {
		t.Errorf("Content-Type = %q, want %q", got.Header().Get("Content-Type"), want)
	}
}

// The dashboard counts rows, so it needs a real database now.
func TestHandleHome(t *testing.T) {
	db := newCustomerTestStore(t)
	customer := newCustomer(t, db)

	got := do(t, db, http.MethodGet, "/")

	if got.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
	}
	if want := "text/html; charset=utf-8"; got.Header().Get("Content-Type") != want {
		t.Errorf("Content-Type = %q, want %q", got.Header().Get("Content-Type"), want)
	}

	// Asserting on a few landmarks rather than the whole document: a test that
	// pins every byte of markup fails on any styling change, which trains you to
	// update it without reading it.
	//
	// The two asset tags matter most — they prove the URLs from the manifest reach
	// the page, which is the one thing that silently breaks the whole frontend.
	body := got.Body.String()
	for _, want := range []string{
		"<!doctype html>",
		"<title>Home · Car Service</title>",
		`<link rel="stylesheet" href="/build/app.css">`,
		`<script type="module" src="/build/app.js">`,
		"Workshop overview",
		// The dashboard shows real counts and the most recent customers.
		"Recent customers",
		customer.Name,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
}

func TestUnknownRouteIsNotFound(t *testing.T) {
	got := do(t, newCustomerTestStore(t), http.MethodGet, "/no-such-page")

	if got.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", got.Code, http.StatusNotFound)
	}
}

// Two things here are deliberately not tested:
//
// Run blocks until the process is killed, so there is no clean way to assert on
// it. What it does — bind a port and serve s.router — is covered by the router
// tests above plus actually running the program.
//
// Serving the bundle reads "public/build" relative to the working directory, and
// `go test` runs each package in its own directory, so the path does not resolve
// here. It is verified by requesting the files from the running server instead.
