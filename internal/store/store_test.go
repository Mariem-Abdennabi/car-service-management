package store

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

// newTestStore connects to the database named by DATABASE_URL.
//
// The test is skipped when that variable is not set, so `go test ./...` still
// passes on a machine with no database — a test suite that cannot run at all is
// worse than one that says why it did not.
func newTestStore(t *testing.T) *Store {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping database tests")
	}

	s, err := New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	t.Cleanup(s.Close)

	return s
}

// createTestCustomer inserts a customer and removes it when the test finishes, so
// tests leave the database as they found it and can be run repeatedly.
func createTestCustomer(t *testing.T, s *Store) Customer {
	t.Helper()

	ctx := context.Background()

	customer, err := s.CreateCustomer(ctx, "Test Customer", "+216 00 000 000", "Testville")
	if err != nil {
		t.Fatalf("CreateCustomer() failed: %v", err)
	}

	t.Cleanup(func() {
		if _, err := s.pool.Exec(ctx, `DELETE FROM customers WHERE id = $1`, customer.ID); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})

	return customer
}

func TestCreateCustomer(t *testing.T) {
	s := newTestStore(t)
	got := createTestCustomer(t, s)

	if got.ID == 0 {
		t.Error("ID = 0, want the id the database assigned")
	}
	if got.Name != "Test Customer" {
		t.Errorf("Name = %q, want %q", got.Name, "Test Customer")
	}
	// Proves the column default was applied and read back, not just that the
	// insert succeeded.
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero, want the database default")
	}
}

func TestCustomer(t *testing.T) {
	s := newTestStore(t)
	created := createTestCustomer(t, s)

	t.Run("returns the customer", func(t *testing.T) {
		got, err := s.Customer(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("Customer() failed: %v", err)
		}
		if got.ID != created.ID || got.Name != created.Name {
			t.Errorf("Customer() = %+v, want %+v", got, created)
		}
	})

	t.Run("reports a missing row as pgx.ErrNoRows", func(t *testing.T) {
		_, err := s.Customer(context.Background(), -1)
		// errors.Is looks through the fmt.Errorf wrapping, which is why the store
		// wraps with %w. This is what lets a handler answer 404 instead of 500.
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Errorf("Customer() error = %v, want pgx.ErrNoRows", err)
		}
	})
}

func TestCustomers(t *testing.T) {
	s := newTestStore(t)
	created := createTestCustomer(t, s)

	customers, err := s.Customers(context.Background())
	if err != nil {
		t.Fatalf("Customers() failed: %v", err)
	}

	for _, c := range customers {
		if c.ID == created.ID {
			return
		}
	}
	t.Errorf("Customers() returned %d rows, none of them the one just created", len(customers))
}
