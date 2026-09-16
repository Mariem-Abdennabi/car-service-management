package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/Mariem-Abdennabi/car-service-management/internal/store"
)

// These pages read the database, so they need a real one. Skipped without
// DATABASE_URL, like the store's own tests.
func newCustomerTestStore(t *testing.T) *store.Store {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping database tests")
	}

	db, err := store.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(db.Close)

	return db
}

func TestHandleCustomers(t *testing.T) {
	db := newCustomerTestStore(t)

	// The list renders whatever is in the table, so the assertion has to hold both
	// when it is empty and when it is not.
	got := do(t, db, http.MethodGet, "/customers")

	if got.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
	}

	body := got.Body.String()
	if !strings.Contains(body, "<title>Customers · Car Service</title>") {
		t.Error("body is not the customers page")
	}

	customers, err := db.Customers(context.Background(), "")
	if err != nil {
		t.Fatalf("Customers() failed: %v", err)
	}
	if len(customers) == 0 {
		if !strings.Contains(body, "No customers found") {
			t.Error("table is empty but the page does not show the empty state")
		}

		return
	}
	if !strings.Contains(body, customers[0].Name) {
		t.Errorf("body does not contain %q", customers[0].Name)
	}
}

func TestHandleCustomer(t *testing.T) {
	db := newCustomerTestStore(t)

	created, err := db.CreateCustomer(context.Background(), "Page Test", "+216 00 000 000", "Testville")
	if err != nil {
		t.Fatalf("CreateCustomer() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.DeleteCustomer(context.Background(), created.ID); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})

	t.Run("renders the customer", func(t *testing.T) {
		got := do(t, db, http.MethodGet, fmt.Sprintf("/customers/%d", created.ID))

		if got.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
		}
		body := got.Body.String()
		for _, want := range []string{created.Name, created.Phone, created.City} {
			if !strings.Contains(body, want) {
				t.Errorf("body does not contain %q", want)
			}
		}
	})

	t.Run("answers 404 for an id that does not exist", func(t *testing.T) {
		got := do(t, db, http.MethodGet, "/customers/999999999")

		if got.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", got.Code, http.StatusNotFound)
		}
	})

	// /customers/abc is not a page, so 404 rather than 400.
	t.Run("answers 404 for a non-numeric id", func(t *testing.T) {
		got := do(t, db, http.MethodGet, "/customers/abc")

		if got.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", got.Code, http.StatusNotFound)
		}
	})
}

// post submits a form the way a browser does.
func post(t *testing.T, db *store.Store, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	newTestServer(db).router.ServeHTTP(recorder, req)

	return recorder
}

func TestHandleNewCustomer(t *testing.T) {
	got := do(t, newCustomerTestStore(t), http.MethodGet, "/customers/new")

	if got.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
	}
	for _, want := range []string{`action="/customers"`, `method="post"`, `name="name"`, `name="phone"`, `name="city"`} {
		if !strings.Contains(got.Body.String(), want) {
			t.Errorf("body does not contain %q", want)
		}
	}
}

func TestHandleCreateCustomer(t *testing.T) {
	db := newCustomerTestStore(t)

	t.Run("creates the customer and redirects to it", func(t *testing.T) {
		got := post(t, db, "/customers", url.Values{
			// Padded, to prove the handler trims before storing.
			"name":  {"  Amira Trabelsi  "},
			"phone": {"+216 55 903 214"},
			"city":  {"Sfax"},
		})

		if got.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusSeeOther)
		}

		location := got.Header().Get("Location")
		var id int64
		if _, err := fmt.Sscanf(location, "/customers/%d", &id); err != nil {
			t.Fatalf("Location = %q, want /customers/<id>", location)
		}
		t.Cleanup(func() {
			if err := db.DeleteCustomer(context.Background(), id); err != nil {
				t.Errorf("cleanup failed: %v", err)
			}
		})

		stored, err := db.Customer(context.Background(), id)
		if err != nil {
			t.Fatalf("Customer() failed: %v", err)
		}
		if stored.Name != "Amira Trabelsi" {
			t.Errorf("Name = %q, want it trimmed", stored.Name)
		}
	})

	t.Run("rejects an empty submission and reports every field", func(t *testing.T) {
		got := post(t, db, "/customers", url.Values{})

		if got.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusUnprocessableEntity)
		}
		for _, want := range []string{"Name is required.", "Phone is required.", "City is required."} {
			if !strings.Contains(got.Body.String(), want) {
				t.Errorf("body does not contain %q", want)
			}
		}
	})

	// The point of redrawing the form rather than redirecting: what was typed
	// survives the rejection.
	t.Run("keeps the values that were entered", func(t *testing.T) {
		got := post(t, db, "/customers", url.Values{
			"name":  {"Youssef Gharbi"},
			"phone": {"   "},
			"city":  {"Sousse"},
		})

		if got.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusUnprocessableEntity)
		}

		body := got.Body.String()
		for _, want := range []string{`value="Youssef Gharbi"`, `value="Sousse"`, "Phone is required."} {
			if !strings.Contains(body, want) {
				t.Errorf("body does not contain %q", want)
			}
		}
	})
}

// newCustomer inserts a customer for a test to work on and removes it afterwards.
func newCustomer(t *testing.T, db *store.Store) store.Customer {
	t.Helper()

	customer, err := db.CreateCustomer(context.Background(), "Salma Bouazizi", "+216 22 318 605", "Ariana")
	if err != nil {
		t.Fatalf("CreateCustomer() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.DeleteCustomer(context.Background(), customer.ID); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})

	return customer
}

// insertTestVehicle gives a customer a vehicle, removed when the test finishes.
// The server package has no CreateVehicle to call until step 4b.
func insertTestVehicle(t *testing.T, db *store.Store, customerID int64) error {
	t.Helper()

	vehicle, err := db.CreateVehicle(context.Background(), customerID, "123 TUN 456", "Renault", "Clio", 2019)
	if err != nil {
		return err
	}

	t.Cleanup(func() {
		if err := db.DeleteVehicle(context.Background(), vehicle.ID); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})

	return nil
}

func TestHandleEditCustomer(t *testing.T) {
	db := newCustomerTestStore(t)
	customer := newCustomer(t, db)

	got := do(t, db, http.MethodGet, fmt.Sprintf("/customers/%d/edit", customer.ID))

	if got.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
	}

	// The form must arrive pre-filled, or editing means retyping everything.
	body := got.Body.String()
	for _, want := range []string{
		fmt.Sprintf(`action="/customers/%d"`, customer.ID),
		fmt.Sprintf(`value="%s"`, customer.Name),
		fmt.Sprintf(`value="%s"`, customer.City),
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
}

func TestHandleUpdateCustomer(t *testing.T) {
	db := newCustomerTestStore(t)

	t.Run("saves the changes and redirects", func(t *testing.T) {
		customer := newCustomer(t, db)

		got := post(t, db, fmt.Sprintf("/customers/%d", customer.ID), url.Values{
			"name":  {"Salma Bouazizi-Khelifi"},
			"phone": {customer.Phone},
			"city":  {"Tunis"},
		})

		if got.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusSeeOther)
		}
		if want := fmt.Sprintf("/customers/%d", customer.ID); got.Header().Get("Location") != want {
			t.Errorf("Location = %q, want %q", got.Header().Get("Location"), want)
		}

		stored, err := db.Customer(context.Background(), customer.ID)
		if err != nil {
			t.Fatalf("Customer() failed: %v", err)
		}
		if stored.Name != "Salma Bouazizi-Khelifi" || stored.City != "Tunis" {
			t.Errorf("stored = %+v, want the updated name and city", stored)
		}
	})

	t.Run("rejects an invalid submission without saving", func(t *testing.T) {
		customer := newCustomer(t, db)

		got := post(t, db, fmt.Sprintf("/customers/%d", customer.ID), url.Values{
			"name":  {""},
			"phone": {customer.Phone},
			"city":  {customer.City},
		})

		if got.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusUnprocessableEntity)
		}
		if !strings.Contains(got.Body.String(), "Name is required.") {
			t.Error("body does not report the missing name")
		}

		stored, err := db.Customer(context.Background(), customer.ID)
		if err != nil {
			t.Fatalf("Customer() failed: %v", err)
		}
		if stored.Name != customer.Name {
			t.Errorf("Name = %q, want it unchanged at %q", stored.Name, customer.Name)
		}
	})

	t.Run("answers 404 for an id that does not exist", func(t *testing.T) {
		got := post(t, db, "/customers/999999999", url.Values{
			"name": {"X"}, "phone": {"Y"}, "city": {"Z"},
		})

		if got.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", got.Code, http.StatusNotFound)
		}
	})
}

func TestHandleDeleteCustomer(t *testing.T) {
	db := newCustomerTestStore(t)

	t.Run("removes the customer and returns to the list", func(t *testing.T) {
		customer, err := db.CreateCustomer(context.Background(), "To Delete", "+216 00 000 000", "Nowhere")
		if err != nil {
			t.Fatalf("CreateCustomer() failed: %v", err)
		}

		got := post(t, db, fmt.Sprintf("/customers/%d/delete", customer.ID), url.Values{})

		if got.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusSeeOther)
		}
		if want := "/customers"; got.Header().Get("Location") != want {
			t.Errorf("Location = %q, want %q", got.Header().Get("Location"), want)
		}

		if _, err := db.Customer(context.Background(), customer.ID); !errors.Is(err, store.ErrNotFound) {
			t.Errorf("Customer() error = %v, want ErrNotFound after deleting", err)
		}
	})

	t.Run("answers 404 for an id that does not exist", func(t *testing.T) {
		got := post(t, db, "/customers/999999999/delete", url.Values{})

		if got.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", got.Code, http.StatusNotFound)
		}
	})
}

// hxGet sends a GET the way htmx does, with the header that marks it as a partial
// request.
func hxGet(t *testing.T, db *store.Store, path string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("HX-Request", "true")

	recorder := httptest.NewRecorder()
	newTestServer(db).router.ServeHTTP(recorder, req)

	return recorder
}

func TestCustomerSearch(t *testing.T) {
	db := newCustomerTestStore(t)
	match := newCustomer(t, db) // Salma Bouazizi, Ariana

	other, err := db.CreateCustomer(context.Background(), "Karim Jebali", "+216 71 260 449", "Bizerte")
	if err != nil {
		t.Fatalf("CreateCustomer() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.DeleteCustomer(context.Background(), other.ID); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})

	t.Run("filters by name", func(t *testing.T) {
		body := do(t, db, http.MethodGet, "/customers?q=salma").Body.String()

		// Lower case on purpose: the query uses ILIKE.
		if !strings.Contains(body, match.Name) {
			t.Errorf("body does not contain %q", match.Name)
		}
		if strings.Contains(body, other.Name) {
			t.Errorf("body contains %q, which does not match the search", other.Name)
		}
	})

	t.Run("matches part of a name", func(t *testing.T) {
		if !strings.Contains(do(t, db, http.MethodGet, "/customers?q=ouaz").Body.String(), match.Name) {
			t.Errorf("body does not contain %q", match.Name)
		}
	})

	t.Run("an empty search returns everyone", func(t *testing.T) {
		body := do(t, db, http.MethodGet, "/customers?q=").Body.String()

		for _, want := range []string{match.Name, other.Name} {
			if !strings.Contains(body, want) {
				t.Errorf("body does not contain %q", want)
			}
		}
	})

	t.Run("says so when nothing matches", func(t *testing.T) {
		if !strings.Contains(do(t, db, http.MethodGet, "/customers?q=zzzznobody").Body.String(), "No customers found") {
			t.Error("body does not report an empty result")
		}
	})

	// The fragment is the whole point: htmx replaces one element, so the response
	// must be that element and nothing else.
	t.Run("an htmx request returns only the list", func(t *testing.T) {
		got := hxGet(t, db, "/customers?q=salma")

		if got.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
		}

		body := got.Body.String()
		if !strings.Contains(body, `id="customer-list"`) {
			t.Error("fragment is missing the id htmx targets, so the next search would have nothing to replace")
		}
		if !strings.Contains(body, match.Name) {
			t.Errorf("fragment does not contain %q", match.Name)
		}
		for _, unwanted := range []string{"<html", "<head", "New customer", "hx-get"} {
			if strings.Contains(body, unwanted) {
				t.Errorf("fragment contains %q, so it is a page rather than a fragment", unwanted)
			}
		}
	})
}

// The confirm step is browser behaviour, so this only checks that the attributes
// reach the page — and, more usefully, that the button is still a real submit
// inside a real form, which is what makes delete work with JavaScript off.
func TestDeleteConfirmIsProgressive(t *testing.T) {
	db := newCustomerTestStore(t)
	customer := newCustomer(t, db)

	body := do(t, db, http.MethodGet, fmt.Sprintf("/customers/%d", customer.ID)).Body.String()

	for _, want := range []string{
		fmt.Sprintf(`action="/customers/%d/delete"`, customer.ID),
		`method="post"`,
		`type="submit"`,
		`x-data="{ confirming: false }"`,
		"x-on:submit=",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("delete form is missing %q", want)
		}
	}
}

// Deleting a customer who still has vehicles is refused by the foreign key. The
// visitor should be told why, not shown a 500.
func TestHandleDeleteCustomerWithVehicles(t *testing.T) {
	db := newCustomerTestStore(t)
	customer := newCustomer(t, db)

	if err := insertTestVehicle(t, db, customer.ID); err != nil {
		t.Fatalf("insert vehicle: %v", err)
	}

	got := post(t, db, fmt.Sprintf("/customers/%d/delete", customer.ID), url.Values{})

	if got.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusConflict)
	}
	if !strings.Contains(got.Body.String(), "still has vehicles") {
		t.Error("body does not explain why the delete was refused")
	}

	if _, err := db.Customer(context.Background(), customer.ID); err != nil {
		t.Errorf("customer was deleted anyway: %v", err)
	}
}
