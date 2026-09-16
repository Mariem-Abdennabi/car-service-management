package store

import (
	"context"
	"errors"
	"os"
	"testing"
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

	t.Run("reports a missing row as ErrNotFound", func(t *testing.T) {
		_, err := s.Customer(context.Background(), -1)
		// The store translates the driver's pgx.ErrNoRows into its own sentinel, so
		// a handler can answer 404 without importing pgx.
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("Customer() error = %v, want ErrNotFound", err)
		}
	})
}

func TestCustomers(t *testing.T) {
	s := newTestStore(t)
	created := createTestCustomer(t, s)

	contains := func(t *testing.T, customers []Customer, id int64) bool {
		t.Helper()

		for _, c := range customers {
			if c.ID == id {
				return true
			}
		}
		return false
	}

	t.Run("an empty search returns everything", func(t *testing.T) {
		customers, err := s.Customers(context.Background(), "")
		if err != nil {
			t.Fatalf("Customers() failed: %v", err)
		}
		if !contains(t, customers, created.ID) {
			t.Errorf("returned %d rows, none of them the one just created", len(customers))
		}
	})

	t.Run("a search filters case-insensitively", func(t *testing.T) {
		customers, err := s.Customers(context.Background(), "test cust")
		if err != nil {
			t.Fatalf("Customers() failed: %v", err)
		}
		if !contains(t, customers, created.ID) {
			t.Error("search did not find the customer by a lower-case fragment of its name")
		}
	})

	t.Run("a search that matches nothing returns nothing", func(t *testing.T) {
		customers, err := s.Customers(context.Background(), "zzzznobody")
		if err != nil {
			t.Fatalf("Customers() failed: %v", err)
		}
		if len(customers) != 0 {
			t.Errorf("returned %d rows, want 0", len(customers))
		}
	})
}
