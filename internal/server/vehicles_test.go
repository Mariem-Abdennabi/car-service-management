package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Mariem-Abdennabi/car-service-management/internal/store"
)

func TestHandleNewVehicle(t *testing.T) {
	db := newCustomerTestStore(t)
	customer := newCustomer(t, db)

	got := do(t, db, http.MethodGet, fmt.Sprintf("/customers/%d/vehicles/new", customer.ID))

	if got.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
	}

	body := got.Body.String()
	for _, want := range []string{
		fmt.Sprintf(`action="/customers/%d/vehicles"`, customer.ID),
		`name="plate"`,
		`type="number"`,
		customer.Name,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
}

func TestHandleCreateVehicle(t *testing.T) {
	db := newCustomerTestStore(t)

	valid := func() url.Values {
		return url.Values{
			"plate": {"123 TUN 456"},
			"make":  {"Renault"},
			"model": {"Clio"},
			"year":  {"2019"},
		}
	}

	t.Run("adds the vehicle and returns to the customer", func(t *testing.T) {
		customer := newCustomer(t, db)

		got := post(t, db, fmt.Sprintf("/customers/%d/vehicles", customer.ID), valid())

		if got.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusSeeOther)
		}
		if want := fmt.Sprintf("/customers/%d", customer.ID); got.Header().Get("Location") != want {
			t.Errorf("Location = %q, want %q", got.Header().Get("Location"), want)
		}

		vehicles, err := db.VehiclesByCustomer(context.Background(), customer.ID)
		if err != nil {
			t.Fatalf("VehiclesByCustomer() failed: %v", err)
		}
		if len(vehicles) != 1 {
			t.Fatalf("customer has %d vehicles, want 1", len(vehicles))
		}
		if vehicles[0].Year != 2019 || vehicles[0].Plate != "123 TUN 456" {
			t.Errorf("stored = %+v, want the submitted values", vehicles[0])
		}
		t.Cleanup(func() {
			if err := db.DeleteVehicle(context.Background(), vehicles[0].ID); err != nil {
				t.Errorf("cleanup failed: %v", err)
			}
		})
	})

	// The year is the first field that is not a string, so it has three ways to be
	// wrong rather than one.
	t.Run("rejects a bad year", func(t *testing.T) {
		customer := newCustomer(t, db)

		cases := map[string]string{
			"":           "Year is required.",
			"not a year": "Year must be a number.",
			"219":        "Year must be between",
			"2999":       "Year must be between",
		}

		for year, want := range cases {
			t.Run(fmt.Sprintf("year=%q", year), func(t *testing.T) {
				form := valid()
				form.Set("year", year)

				got := post(t, db, fmt.Sprintf("/customers/%d/vehicles", customer.ID), form)

				if got.Code != http.StatusUnprocessableEntity {
					t.Fatalf("status = %d, want %d", got.Code, http.StatusUnprocessableEntity)
				}
				if !strings.Contains(got.Body.String(), want) {
					t.Errorf("body does not contain %q", want)
				}
			})
		}
	})

	t.Run("accepts next year, since new models are sold early", func(t *testing.T) {
		customer := newCustomer(t, db)

		form := valid()
		form.Set("year", fmt.Sprint(time.Now().Year()+1))

		if got := post(t, db, fmt.Sprintf("/customers/%d/vehicles", customer.ID), form); got.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusSeeOther)
		}

		vehicles, err := db.VehiclesByCustomer(context.Background(), customer.ID)
		if err != nil {
			t.Fatalf("VehiclesByCustomer() failed: %v", err)
		}
		t.Cleanup(func() {
			if err := db.DeleteVehicle(context.Background(), vehicles[0].ID); err != nil {
				t.Errorf("cleanup failed: %v", err)
			}
		})
	})

	t.Run("keeps what was typed when something is wrong", func(t *testing.T) {
		customer := newCustomer(t, db)

		form := valid()
		form.Set("plate", "")

		got := post(t, db, fmt.Sprintf("/customers/%d/vehicles", customer.ID), form)

		body := got.Body.String()
		for _, want := range []string{`value="Renault"`, `value="Clio"`, `value="2019"`, "Plate is required."} {
			if !strings.Contains(body, want) {
				t.Errorf("body does not contain %q", want)
			}
		}
	})

	t.Run("answers 404 for a customer that does not exist", func(t *testing.T) {
		if got := post(t, db, "/customers/999999999/vehicles", valid()); got.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", got.Code, http.StatusNotFound)
		}
	})
}

// newVehicle gives a customer a vehicle for a test to work on.
func newVehicle(t *testing.T, db *store.Store, customerID int64) store.Vehicle {
	t.Helper()

	vehicle, err := db.CreateVehicle(context.Background(), customerID, "123 TUN 456", "Renault", "Clio", 2019)
	if err != nil {
		t.Fatalf("CreateVehicle() failed: %v", err)
	}
	t.Cleanup(func() {
		// Already gone in the delete tests, and deleting nothing is not an error.
		if err := db.DeleteVehicle(context.Background(), vehicle.ID); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})

	return vehicle
}

func TestHandleEditVehicle(t *testing.T) {
	db := newCustomerTestStore(t)
	customer := newCustomer(t, db)
	vehicle := newVehicle(t, db, customer.ID)

	got := do(t, db, http.MethodGet, fmt.Sprintf("/vehicles/%d/edit", vehicle.ID))

	if got.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
	}

	body := got.Body.String()
	for _, want := range []string{
		fmt.Sprintf(`action="/vehicles/%d"`, vehicle.ID),
		`value="123 TUN 456"`,
		`value="2019"`,
		// The owner is read from the vehicle, so their name appears without the URL
		// having mentioned them.
		customer.Name,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
}

func TestHandleUpdateVehicle(t *testing.T) {
	db := newCustomerTestStore(t)

	t.Run("saves the changes and returns to the customer", func(t *testing.T) {
		customer := newCustomer(t, db)
		vehicle := newVehicle(t, db, customer.ID)

		got := post(t, db, fmt.Sprintf("/vehicles/%d", vehicle.ID), url.Values{
			"plate": {"789 TUN 012"},
			"make":  {"Peugeot"},
			"model": {"208"},
			"year":  {"2021"},
		})

		if got.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusSeeOther)
		}
		if want := fmt.Sprintf("/customers/%d", customer.ID); got.Header().Get("Location") != want {
			t.Errorf("Location = %q, want %q", got.Header().Get("Location"), want)
		}

		stored, err := db.Vehicle(context.Background(), vehicle.ID)
		if err != nil {
			t.Fatalf("Vehicle() failed: %v", err)
		}
		if stored.Plate != "789 TUN 012" || stored.Year != 2021 {
			t.Errorf("stored = %+v, want the updated values", stored)
		}
	})

	t.Run("rejects an invalid submission without saving", func(t *testing.T) {
		customer := newCustomer(t, db)
		vehicle := newVehicle(t, db, customer.ID)

		got := post(t, db, fmt.Sprintf("/vehicles/%d", vehicle.ID), url.Values{
			"plate": {"789 TUN 012"},
			"make":  {"Peugeot"},
			"model": {"208"},
			"year":  {"nope"},
		})

		if got.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusUnprocessableEntity)
		}

		stored, err := db.Vehicle(context.Background(), vehicle.ID)
		if err != nil {
			t.Fatalf("Vehicle() failed: %v", err)
		}
		if stored.Plate != vehicle.Plate {
			t.Errorf("Plate = %q, want it unchanged at %q", stored.Plate, vehicle.Plate)
		}
	})

	t.Run("answers 404 for a vehicle that does not exist", func(t *testing.T) {
		got := post(t, db, "/vehicles/999999999", url.Values{
			"plate": {"X"}, "make": {"Y"}, "model": {"Z"}, "year": {"2020"},
		})

		if got.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", got.Code, http.StatusNotFound)
		}
	})
}

func TestHandleDeleteVehicle(t *testing.T) {
	db := newCustomerTestStore(t)
	customer := newCustomer(t, db)

	vehicle, err := db.CreateVehicle(context.Background(), customer.ID, "345 TUN 678", "Fiat", "Panda", 2015)
	if err != nil {
		t.Fatalf("CreateVehicle() failed: %v", err)
	}

	got := post(t, db, fmt.Sprintf("/vehicles/%d/delete", vehicle.ID), url.Values{})

	if got.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusSeeOther)
	}
	if want := fmt.Sprintf("/customers/%d", customer.ID); got.Header().Get("Location") != want {
		t.Errorf("Location = %q, want %q", got.Header().Get("Location"), want)
	}

	if _, err := db.Vehicle(context.Background(), vehicle.ID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Vehicle() error = %v, want ErrNotFound after deleting", err)
	}

	// The whole point of ON DELETE RESTRICT: once the vehicles are gone, the
	// customer can be deleted again.
	if err := db.DeleteCustomer(context.Background(), customer.ID); err != nil {
		t.Errorf("DeleteCustomer() failed after removing the vehicle: %v", err)
	}
}
