package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Mariem-Abdennabi/car-service-management/internal/store"
)

// newJobFor gives a customer a vehicle with one job on it, cleaned up afterwards.
//
// The cleanups are registered vehicle-first, job-second, so they run in the
// opposite order: t.Cleanup is last-in-first-out, and neither a vehicle with jobs
// nor a customer with vehicles can be deleted.
func newJobFor(t *testing.T, db *store.Store, customer store.Customer) (store.Vehicle, store.Job) {
	t.Helper()

	ctx := context.Background()

	vehicle, err := db.CreateVehicle(ctx, customer.ID, "123 TUN 4567", "Renault", "Clio", 2019)
	if err != nil {
		t.Fatalf("CreateVehicle() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.DeleteVehicle(ctx, vehicle.ID); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})

	job, err := db.CreateJob(ctx, vehicle.ID, "Grinding noise when braking", store.JobInProgress, 45000)
	if err != nil {
		t.Fatalf("CreateJob() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.DeleteJob(ctx, job.ID); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})

	return vehicle, job
}

func TestHandleJobs(t *testing.T) {
	db := newCustomerTestStore(t)
	customer := newCustomer(t, db)
	vehicle, job := newJobFor(t, db, customer)

	got := do(t, db, http.MethodGet, "/jobs")

	if got.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
	}

	// The list joins three tables, so the assertion is that all three arrived: the
	// job, the car it is on, and whose car it is.
	body := got.Body.String()
	for _, want := range []string{
		"<title>Service jobs · Car Service</title>",
		job.Description,
		vehicle.Plate,
		customer.Name,
		// The stored status is never what is shown.
		"In progress",
		fmt.Sprintf(`href="/jobs/%d"`, job.ID),
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
	if strings.Contains(body, store.JobInProgress) {
		t.Errorf("body contains the raw status %q, want the label only", store.JobInProgress)
	}
}

func TestHandleJob(t *testing.T) {
	db := newCustomerTestStore(t)
	customer := newCustomer(t, db)
	vehicle, job := newJobFor(t, db, customer)

	got := do(t, db, http.MethodGet, fmt.Sprintf("/jobs/%d", job.ID))

	if got.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
	}

	body := got.Body.String()
	for _, want := range []string{
		job.Description,
		vehicle.Plate,
		// The owner is reached by following the job to its vehicle, so this proves
		// that chain works rather than that the URL was trusted.
		customer.Name,
		fmt.Sprintf(`href="/customers/%d"`, customer.ID),
		// Labour, in dinars rather than millimes.
		"45.000 TND",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
}

func TestHandleJobNotFound(t *testing.T) {
	db := newCustomerTestStore(t)

	// A job that does not exist and a path that is not a number both reach the same
	// page: an id the application cannot resolve is a missing page either way.
	for _, path := range []string{"/jobs/0", "/jobs/not-a-number"} {
		got := do(t, db, http.MethodGet, path)

		if got.Code != http.StatusNotFound {
			t.Errorf("GET %s status = %d, want %d", path, got.Code, http.StatusNotFound)
		}
	}
}

func TestCustomerPageShowsJobs(t *testing.T) {
	db := newCustomerTestStore(t)
	customer := newCustomer(t, db)
	_, job := newJobFor(t, db, customer)

	got := do(t, db, http.MethodGet, fmt.Sprintf("/customers/%d", customer.ID))

	if got.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
	}

	body := got.Body.String()
	for _, want := range []string{"Service jobs", job.Description} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
}

// Deleting a vehicle that has been worked on is refused and explained, rather
// than becoming a 500 the moment service_jobs started referencing vehicles.
func TestHandleDeleteVehicleWithJobs(t *testing.T) {
	db := newCustomerTestStore(t)
	customer := newCustomer(t, db)
	vehicle, _ := newJobFor(t, db, customer)

	got := post(t, db, fmt.Sprintf("/vehicles/%d/delete", vehicle.ID), nil)

	if got.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusConflict)
	}
	if want := "This vehicle has service jobs"; !strings.Contains(got.Body.String(), want) {
		t.Errorf("body does not contain %q", want)
	}
}
