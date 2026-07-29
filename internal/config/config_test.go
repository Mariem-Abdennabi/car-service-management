package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr bool
	}{
		{
			name: "applies defaults for the optional settings",
			env:  map[string]string{"DATABASE_URL": "postgres://localhost:5432/db"},
			want: Config{Env: "development", Port: 8080, DatabaseURL: "postgres://localhost:5432/db"},
		},
		{
			name: "reads values from the environment",
			env: map[string]string{
				"APP_ENV":      "production",
				"PORT":         "3000",
				"DATABASE_URL": "postgres://localhost:5432/db",
			},
			want: Config{Env: "production", Port: 3000, DatabaseURL: "postgres://localhost:5432/db"},
		},
		{
			name:    "rejects an unknown environment name",
			env:     map[string]string{"APP_ENV": "staging", "DATABASE_URL": "postgres://localhost/db"},
			wantErr: true,
		},
		{
			name:    "rejects a non-numeric port",
			env:     map[string]string{"PORT": "eighty-eighty", "DATABASE_URL": "postgres://localhost/db"},
			wantErr: true,
		},
		{
			name:    "rejects a port outside the valid range",
			env:     map[string]string{"PORT": "70000", "DATABASE_URL": "postgres://localhost/db"},
			wantErr: true,
		},
		{
			name:    "rejects a missing DATABASE_URL",
			env:     map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Every variable is set explicitly, so a value left over from the
			// developer's own shell cannot change the result. t.Setenv restores
			// the previous value when the subtest finishes.
			for _, key := range []string{"APP_ENV", "PORT", "DATABASE_URL"} {
				t.Setenv(key, tt.env[key])
			}

			got, err := Load()

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() succeeded with %+v, want an error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() returned an unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	t.Run("fills in variables that are not already set", func(t *testing.T) {
		path := writeEnvFile(t, "APP_ENV=production\nPORT=3000\n")
		unset(t, "APP_ENV", "PORT")

		if err := LoadFile(path); err != nil {
			t.Fatalf("LoadFile() returned an unexpected error: %v", err)
		}
		if got := os.Getenv("APP_ENV"); got != "production" {
			t.Errorf("APP_ENV = %q, want %q", got, "production")
		}
		if got := os.Getenv("PORT"); got != "3000" {
			t.Errorf("PORT = %q, want %q", got, "3000")
		}
	})

	t.Run("leaves an already-set variable alone", func(t *testing.T) {
		path := writeEnvFile(t, "PORT=3000\n")
		t.Setenv("PORT", "9999")

		if err := LoadFile(path); err != nil {
			t.Fatalf("LoadFile() returned an unexpected error: %v", err)
		}
		if got := os.Getenv("PORT"); got != "9999" {
			t.Errorf("PORT = %q, want the exported value %q to win", got, "9999")
		}
	})

	t.Run("treats a missing file as no configuration", func(t *testing.T) {
		if err := LoadFile(filepath.Join(t.TempDir(), "does-not-exist")); err != nil {
			t.Errorf("LoadFile() on a missing file returned %v, want nil", err)
		}
	})
}

// writeEnvFile puts contents into a .env file inside a temporary directory that
// the test framework removes afterwards, and returns its path.
func writeEnvFile(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write test .env: %v", err)
	}
	return path
}

// unset removes environment variables for the duration of the test.
//
// t.Setenv is called first purely for its cleanup: it registers a restore of the
// original value when the test ends. os.Unsetenv then makes the variable truly
// absent, which matters because godotenv only fills in variables that are not
// present at all — one set to the empty string counts as present.
func unset(t *testing.T, keys ...string) {
	t.Helper()

	for _, key := range keys {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}
}

func TestIsDevelopment(t *testing.T) {
	if !(Config{Env: "development"}).IsDevelopment() {
		t.Error("IsDevelopment() = false for development, want true")
	}
	if (Config{Env: "production"}).IsDevelopment() {
		t.Error("IsDevelopment() = true for production, want false")
	}
}
