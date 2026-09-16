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
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Mariem-Abdennabi/car-service-management/internal/db"
)

// ErrNotFound is returned when a row does not exist.
//
// The driver's own pgx.ErrNoRows is translated to this so that handlers can answer
// 404 without importing pgx. Layers above this one should not have to know which
// database library is underneath.
var ErrNotFound = errors.New("not found")

// ErrInUse is returned when a row cannot be deleted because another row still
// references it — a customer who still has vehicles.
//
// PostgreSQL reports this as SQLSTATE 23503, a foreign key violation. Translating
// it here means the handler can explain the problem instead of answering 500.
var ErrInUse = errors.New("still referenced")

// ErrDuplicate is returned when a value has to be unique and already exists — two
// parts with the same reference.
//
// PostgreSQL reports this as SQLSTATE 23505. Translating it means the form can
// point at the offending field instead of the request becoming a 500.
var ErrDuplicate = errors.New("already exists")

// Customer is someone who brings a vehicle in for service.
//
// An alias, not a copy: `store.Customer` *is* `db.Customer`. That keeps callers
// out of the generated package without a hand-written struct and a mapping
// function to keep in step with the schema.
type Customer = db.Customer

// Vehicle is a car belonging to a customer.
type Vehicle = db.Vehicle

// Part is a spare part in the workshop's catalogue.
type Part = db.Part

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

// Customers returns customers whose name contains search, newest first. An empty
// search returns all of them.
func (s *Store) Customers(ctx context.Context, search string) ([]Customer, error) {
	customers, err := s.queries.ListCustomers(ctx, search)
	if err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}

	return customers, nil
}

// Customer returns one customer by id, or ErrNotFound if there is no such row.
func (s *Store) Customer(ctx context.Context, id int64) (Customer, error) {
	customer, err := s.queries.GetCustomer(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Customer{}, ErrNotFound
	}
	if err != nil {
		return Customer{}, fmt.Errorf("get customer %d: %w", id, err)
	}

	return customer, nil
}

// UpdateCustomer overwrites a customer's details, or returns ErrNotFound if there
// is no such row.
func (s *Store) UpdateCustomer(ctx context.Context, id int64, name, phone, city string) (Customer, error) {
	customer, err := s.queries.UpdateCustomer(ctx, db.UpdateCustomerParams{
		ID:    id,
		Name:  name,
		Phone: phone,
		City:  city,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Customer{}, ErrNotFound
	}
	if err != nil {
		return Customer{}, fmt.Errorf("update customer %d: %w", id, err)
	}

	return customer, nil
}

// DeleteCustomer removes a customer. Deleting a row that is not there is not an
// error: the end state the caller asked for is the end state they get.
func (s *Store) DeleteCustomer(ctx context.Context, id int64) error {
	if err := s.queries.DeleteCustomer(ctx, id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
			return ErrInUse
		}

		return fmt.Errorf("delete customer %d: %w", id, err)
	}

	return nil
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

// VehiclesByCustomer returns one customer's vehicles, newest first.
func (s *Store) VehiclesByCustomer(ctx context.Context, customerID int64) ([]Vehicle, error) {
	vehicles, err := s.queries.ListVehiclesByCustomer(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("list vehicles for customer %d: %w", customerID, err)
	}

	return vehicles, nil
}

// Vehicle returns one vehicle by id, or ErrNotFound if there is no such row.
func (s *Store) Vehicle(ctx context.Context, id int64) (Vehicle, error) {
	vehicle, err := s.queries.GetVehicle(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Vehicle{}, ErrNotFound
	}
	if err != nil {
		return Vehicle{}, fmt.Errorf("get vehicle %d: %w", id, err)
	}

	return vehicle, nil
}

// UpdateVehicle overwrites a vehicle's details, or returns ErrNotFound if there is
// no such row.
func (s *Store) UpdateVehicle(ctx context.Context, id int64, plate, make, model string, year int32) (Vehicle, error) {
	vehicle, err := s.queries.UpdateVehicle(ctx, db.UpdateVehicleParams{
		ID:    id,
		Plate: plate,
		Make:  make,
		Model: model,
		Year:  year,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Vehicle{}, ErrNotFound
	}
	if err != nil {
		return Vehicle{}, fmt.Errorf("update vehicle %d: %w", id, err)
	}

	return vehicle, nil
}

// CreateVehicle adds a vehicle to a customer.
func (s *Store) CreateVehicle(ctx context.Context, customerID int64, plate, make, model string, year int32) (Vehicle, error) {
	vehicle, err := s.queries.CreateVehicle(ctx, db.CreateVehicleParams{
		CustomerID: customerID,
		Plate:      plate,
		Make:       make,
		Model:      model,
		Year:       year,
	})
	if err != nil {
		return Vehicle{}, fmt.Errorf("create vehicle: %w", err)
	}

	return vehicle, nil
}

// DeleteVehicle removes a vehicle.
func (s *Store) DeleteVehicle(ctx context.Context, id int64) error {
	if err := s.queries.DeleteVehicle(ctx, id); err != nil {
		return fmt.Errorf("delete vehicle %d: %w", id, err)
	}

	return nil
}

// Parts returns the catalogue, by name.
func (s *Store) Parts(ctx context.Context) ([]Part, error) {
	parts, err := s.queries.ListParts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list parts: %w", err)
	}

	return parts, nil
}

// Part returns one part by id, or ErrNotFound if there is no such row.
func (s *Store) Part(ctx context.Context, id int64) (Part, error) {
	part, err := s.queries.GetPart(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Part{}, ErrNotFound
	}
	if err != nil {
		return Part{}, fmt.Errorf("get part %d: %w", id, err)
	}

	return part, nil
}

// CreatePart adds a part to the catalogue, or returns ErrDuplicate if the
// reference is already taken.
func (s *Store) CreatePart(ctx context.Context, reference, name string, priceMillimes, quantity int32) (Part, error) {
	part, err := s.queries.CreatePart(ctx, db.CreatePartParams{
		Reference:      reference,
		Name:           name,
		PriceMillimes:  priceMillimes,
		QuantityOnHand: quantity,
	})
	if isDuplicate(err) {
		return Part{}, ErrDuplicate
	}
	if err != nil {
		return Part{}, fmt.Errorf("create part: %w", err)
	}

	return part, nil
}

// UpdatePart overwrites a part, returning ErrNotFound if there is no such row or
// ErrDuplicate if the new reference belongs to another part.
func (s *Store) UpdatePart(ctx context.Context, id int64, reference, name string, priceMillimes, quantity int32) (Part, error) {
	part, err := s.queries.UpdatePart(ctx, db.UpdatePartParams{
		ID:             id,
		Reference:      reference,
		Name:           name,
		PriceMillimes:  priceMillimes,
		QuantityOnHand: quantity,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Part{}, ErrNotFound
	}
	if isDuplicate(err) {
		return Part{}, ErrDuplicate
	}
	if err != nil {
		return Part{}, fmt.Errorf("update part %d: %w", id, err)
	}

	return part, nil
}

// isDuplicate reports whether err is PostgreSQL refusing a duplicate value on a
// unique index.
func isDuplicate(err error) bool {
	var pgErr *pgconn.PgError

	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}

// DeletePart removes a part.
func (s *Store) DeletePart(ctx context.Context, id int64) error {
	if err := s.queries.DeletePart(ctx, id); err != nil {
		return fmt.Errorf("delete part %d: %w", id, err)
	}

	return nil
}

// Summary is what the home page shows: how much of everything there is.
type Summary struct {
	Customers  int64
	Vehicles   int64
	Parts      int64
	OutOfStock int64
}

// Summary returns the counts for the home page in one round trip.
func (s *Store) Summary(ctx context.Context) (Summary, error) {
	counts, err := s.queries.Counts(ctx)
	if err != nil {
		return Summary{}, fmt.Errorf("count rows: %w", err)
	}

	return Summary{
		Customers:  counts.Customers,
		Vehicles:   counts.Vehicles,
		Parts:      counts.Parts,
		OutOfStock: counts.OutOfStock,
	}, nil
}

// PartsOutOfStock returns up to five parts with nothing on the shelf.
func (s *Store) PartsOutOfStock(ctx context.Context) ([]Part, error) {
	parts, err := s.queries.ListPartsOutOfStock(ctx)
	if err != nil {
		return nil, fmt.Errorf("list parts out of stock: %w", err)
	}

	return parts, nil
}

// RecentCustomers returns the five most recently added customers.
func (s *Store) RecentCustomers(ctx context.Context) ([]Customer, error) {
	customers, err := s.queries.ListRecentCustomers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list recent customers: %w", err)
	}

	return customers, nil
}
