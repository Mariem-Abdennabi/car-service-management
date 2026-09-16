package store

import (
	"context"
	"errors"
	"testing"
)

// insertVehicle gives a customer a vehicle, removed when the test finishes.
func insertVehicle(t *testing.T, s *Store, customerID int64, plate string) int64 {
	t.Helper()

	ctx := context.Background()

	vehicle, err := s.CreateVehicle(ctx, customerID, plate, "Renault", "Clio", 2019)
	if err != nil {
		t.Fatalf("CreateVehicle() failed: %v", err)
	}

	t.Cleanup(func() {
		if err := s.DeleteVehicle(ctx, vehicle.ID); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})

	return vehicle.ID
}

func TestVehiclesByCustomer(t *testing.T) {
	s := newTestStore(t)
	owner := createTestCustomer(t, s)
	other := createTestCustomer(t, s)

	mine := insertVehicle(t, s, owner.ID, "123 TUN 456")
	insertVehicle(t, s, other.ID, "789 TUN 012")

	vehicles, err := s.VehiclesByCustomer(context.Background(), owner.ID)
	if err != nil {
		t.Fatalf("VehiclesByCustomer() failed: %v", err)
	}

	// The foreign key is only useful if the query respects it: one customer's
	// vehicles must not appear on another's.
	if len(vehicles) != 1 {
		t.Fatalf("returned %d vehicles, want 1", len(vehicles))
	}
	if vehicles[0].ID != mine {
		t.Errorf("returned vehicle %d, want %d", vehicles[0].ID, mine)
	}
}

// The point of ON DELETE RESTRICT: a customer with vehicles cannot be removed, so
// their records cannot be lost by accident. Step 4c turns this into a message.
func TestDeleteCustomerWithVehiclesIsRefused(t *testing.T) {
	s := newTestStore(t)
	customer := createTestCustomer(t, s)
	insertVehicle(t, s, customer.ID, "345 TUN 678")

	if err := s.DeleteCustomer(context.Background(), customer.ID); !errors.Is(err, ErrInUse) {
		t.Errorf("DeleteCustomer() error = %v, want ErrInUse", err)
	}

	if _, err := s.Customer(context.Background(), customer.ID); err != nil {
		t.Errorf("customer was removed anyway: %v", err)
	}
}
