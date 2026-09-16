package store

import (
	"context"
	"errors"
	"testing"
)

// insertJob opens a job on a vehicle, removed when the test finishes.
//
// Registered after the vehicle's own cleanup, so it runs first: t.Cleanup is
// last-in-first-out, and a vehicle with jobs cannot be deleted.
func insertJob(t *testing.T, s *Store, vehicleID int64, description, status string) Job {
	t.Helper()

	ctx := context.Background()

	job, err := s.CreateJob(ctx, vehicleID, description, status, 25000)
	if err != nil {
		t.Fatalf("CreateJob() failed: %v", err)
	}

	t.Cleanup(func() {
		if err := s.DeleteJob(ctx, job.ID); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})

	return job
}

func TestCreateJob(t *testing.T) {
	s := newTestStore(t)
	customer := createTestCustomer(t, s)
	vehicle := insertVehicle(t, s, customer.ID, "123 TUN 456")

	got := insertJob(t, s, vehicle, "Grinding noise when braking", JobReceived)

	if got.ID == 0 {
		t.Error("ID = 0, want the id the database assigned")
	}
	if got.Status != JobReceived {
		t.Errorf("Status = %q, want %q", got.Status, JobReceived)
	}
	// Proves the column default was applied and read back, not just that the insert
	// succeeded.
	if got.OpenedAt.IsZero() {
		t.Error("OpenedAt is zero, want the database default")
	}
}

// The CHECK constraint is the reason status is safe to render without a default
// case in every template: the database will not hold anything else.
func TestCreateJobRejectsAnUnknownStatus(t *testing.T) {
	s := newTestStore(t)
	customer := createTestCustomer(t, s)
	vehicle := insertVehicle(t, s, customer.ID, "123 TUN 456")

	if _, err := s.CreateJob(context.Background(), vehicle, "Anything", "on_fire", 0); err == nil {
		t.Error("CreateJob() succeeded with an unknown status, want the CHECK constraint to refuse it")
	}
}

func TestJobsByCustomer(t *testing.T) {
	s := newTestStore(t)
	owner := createTestCustomer(t, s)
	other := createTestCustomer(t, s)

	ownersVehicle := insertVehicle(t, s, owner.ID, "123 TUN 456")
	othersVehicle := insertVehicle(t, s, other.ID, "789 TUN 012")

	mine := insertJob(t, s, ownersVehicle, "Clutch slipping", JobInProgress)
	insertJob(t, s, othersVehicle, "Air conditioning", JobReceived)

	jobs, err := s.JobsByCustomer(context.Background(), owner.ID)
	if err != nil {
		t.Fatalf("JobsByCustomer() failed: %v", err)
	}

	// The query reaches the customer through the vehicle, so this is what proves
	// the join goes where it should: one customer's jobs must not appear on
	// another's page.
	if len(jobs) != 1 {
		t.Fatalf("returned %d jobs, want 1", len(jobs))
	}
	if jobs[0].ID != mine.ID {
		t.Errorf("returned job %d, want %d", jobs[0].ID, mine.ID)
	}
	if jobs[0].Plate != "123 TUN 456" {
		t.Errorf("Plate = %q, want the vehicle the job is on", jobs[0].Plate)
	}
}

func TestJob(t *testing.T) {
	s := newTestStore(t)
	customer := createTestCustomer(t, s)
	vehicle := insertVehicle(t, s, customer.ID, "123 TUN 456")
	created := insertJob(t, s, vehicle, "Annual service", JobCompleted)

	t.Run("returns the job", func(t *testing.T) {
		got, err := s.Job(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("Job() failed: %v", err)
		}
		if got.ID != created.ID || got.Description != created.Description {
			t.Errorf("Job() = %+v, want %+v", got, created)
		}
	})

	t.Run("reports a missing job as ErrNotFound", func(t *testing.T) {
		if _, err := s.Job(context.Background(), 0); !errors.Is(err, ErrNotFound) {
			t.Errorf("Job() error = %v, want ErrNotFound", err)
		}
	})
}

// The same ON DELETE RESTRICT reasoning as customers and vehicles, one level
// further down: a car that has been worked on keeps its history.
func TestDeleteVehicleWithJobsIsRefused(t *testing.T) {
	s := newTestStore(t)
	customer := createTestCustomer(t, s)
	vehicle := insertVehicle(t, s, customer.ID, "345 TUN 678")
	insertJob(t, s, vehicle, "Timing belt", JobInProgress)

	if err := s.DeleteVehicle(context.Background(), vehicle); !errors.Is(err, ErrInUse) {
		t.Errorf("DeleteVehicle() error = %v, want ErrInUse", err)
	}
}
