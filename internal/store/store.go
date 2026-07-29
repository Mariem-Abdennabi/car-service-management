// Package store is the application's database access.
//
// It owns the connection pool and wraps the code sqlc generates from
// sql/queries/*.sql. Until step 2c the SQL and every rows.Scan call in here were
// written by hand; `git log` on this file shows what that looked like.
//
// The wrapper is thin on purpose. It exists so the rest of the application imports
// one package and gets errors with context attached, rather than importing the
// generated package directly.
package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Mariem-Abdennabi/car-service-management/internal/db"
)

// Customer is someone who brings a vehicle in for service.
//
// An alias, not a copy: `store.Customer` *is* `db.Customer`. That keeps callers
// out of the generated package without a hand-written struct and a mapping
// function to keep in step with the schema.
type Customer = db.Customer

// Store holds the connection pool and the generated queries that use it.
type Store struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// New opens a connection pool and verifies the database is actually reachable.
//
// A pool is lazy: pgxpool.New only parses the URL, so without the Ping a wrong
// password or a stopped server would not surface until the first request. Checking
// here turns that into a startup error.
//
// The caller must call Close when finished.
func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Store{pool: pool, queries: db.New(pool)}, nil
}

// Close releases every connection in the pool.
func (s *Store) Close() {
	s.pool.Close()
}

// Ping reports whether the database is reachable. Used by the health check.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// Customers returns every customer, newest first.
func (s *Store) Customers(ctx context.Context) ([]Customer, error) {
	customers, err := s.queries.ListCustomers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}

	return customers, nil
}

// Customer returns one customer by id.
//
// A missing row comes back as pgx.ErrNoRows, which callers check with
// errors.Is(err, pgx.ErrNoRows) to answer 404 rather than 500.
func (s *Store) Customer(ctx context.Context, id int64) (Customer, error) {
	customer, err := s.queries.GetCustomer(ctx, id)
	if err != nil {
		return Customer{}, fmt.Errorf("get customer %d: %w", id, err)
	}

	return customer, nil
}

// CreateCustomer inserts a customer and returns it with the id and timestamp the
// database assigned.
//
// The generated method takes a params struct, which this builds — so callers keep
// passing plain arguments and did not change when the generated code arrived.
func (s *Store) CreateCustomer(ctx context.Context, name, phone, city string) (Customer, error) {
	customer, err := s.queries.CreateCustomer(ctx, db.CreateCustomerParams{
		Name:  name,
		Phone: phone,
		City:  city,
	})
	if err != nil {
		return Customer{}, fmt.Errorf("create customer: %w", err)
	}

	return customer, nil
}
