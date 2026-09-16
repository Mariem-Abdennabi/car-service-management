package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Mariem-Abdennabi/car-service-management/internal/store"
)

// Test parts carry a reference no real catalogue entry uses, so a test cannot
// collide with the rows `make seed` puts in the same table. The unique index means
// a collision is a failed insert, not a confusing assertion.
const testReference = "TEST-"

// newPart adds a part for a test to work on, removed when the test finishes.
func newPart(t *testing.T, db *store.Store, reference, name string, price, quantity int32) store.Part {
	t.Helper()

	part, err := db.CreatePart(context.Background(), reference, name, price, quantity)
	if err != nil {
		t.Fatalf("CreatePart() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.DeletePart(context.Background(), part.ID); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})

	return part
}

// findPart looks a part up by reference and removes it when the test finishes.
//
// By reference rather than by counting the catalogue: `make seed` fills the parts
// table, so a test that assumes it is empty only passes on a fresh database — and
// worse, it fails *after* creating its row, so the cleanup is never registered and
// the row is orphaned. Every later run then hits the unique index instead.
func findPart(t *testing.T, db *store.Store, reference string) store.Part {
	t.Helper()

	parts, err := db.Parts(context.Background())
	if err != nil {
		t.Fatalf("Parts() failed: %v", err)
	}

	for _, part := range parts {
		if part.Reference != reference {
			continue
		}

		t.Cleanup(func() {
			if err := db.DeletePart(context.Background(), part.ID); err != nil {
				t.Errorf("cleanup failed: %v", err)
			}
		})

		return part
	}

	t.Fatalf("no part with reference %q in the catalogue", reference)

	return store.Part{}
}

func TestHandleParts(t *testing.T) {
	db := newCustomerTestStore(t)

	inStock := newPart(t, db, testReference+"001", "Brake pad set", 42500, 12)
	none := newPart(t, db, testReference+"005", "Oil filter", 8000, 0)

	got := do(t, db, http.MethodGet, "/parts")

	if got.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
	}

	body := got.Body.String()
	for _, want := range []string{
		inStock.Reference,
		inStock.Name,
		// Stored as millimes, shown as dinars.
		"42.500 TND",
		none.Reference,
		"Out of stock",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
}

func TestHandleCreatePart(t *testing.T) {
	db := newCustomerTestStore(t)

	valid := func() url.Values {
		return url.Values{
			"reference": {testReference + "777"},
			"name":      {"Brake disc"},
			"price":     {"42.500"},
			"quantity":  {"6"},
		}
	}

	t.Run("adds the part with the price in millimes", func(t *testing.T) {
		got := post(t, db, "/parts", valid())

		if got.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusSeeOther)
		}

		created := findPart(t, db, testReference+"777")

		// The point of the whole millimes decision: 42.500 dinars is stored exactly.
		if created.PriceMillimes != 42500 {
			t.Errorf("PriceMillimes = %d, want 42500", created.PriceMillimes)
		}
	})

	t.Run("rejects a duplicate reference as a field error", func(t *testing.T) {
		existing := newPart(t, db, testReference+"888", "Brake pad set", 30000, 3)

		form := valid()
		form.Set("reference", existing.Reference)

		got := post(t, db, "/parts", form)

		// 422 with a message on the field, not a 500 from the unique index.
		if got.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusUnprocessableEntity)
		}
		if !strings.Contains(got.Body.String(), "already used by another part") {
			t.Error("body does not explain the duplicate reference")
		}
	})

	t.Run("rejects a bad price and keeps what was typed", func(t *testing.T) {
		form := valid()
		form.Set("price", "42.5000")

		got := post(t, db, "/parts", form)

		if got.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusUnprocessableEntity)
		}
		body := got.Body.String()
		for _, want := range []string{"Price must be an amount in dinars", `value="Brake disc"`, `value="6"`} {
			if !strings.Contains(body, want) {
				t.Errorf("body does not contain %q", want)
			}
		}
	})

	t.Run("rejects a negative quantity", func(t *testing.T) {
		form := valid()
		form.Set("quantity", "-3")

		got := post(t, db, "/parts", form)

		if got.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusUnprocessableEntity)
		}
		if !strings.Contains(got.Body.String(), "cannot be negative") {
			t.Error("body does not reject the negative quantity")
		}
	})
}

func TestHandleEditAndDeletePart(t *testing.T) {
	db := newCustomerTestStore(t)

	t.Run("the edit form round-trips the price", func(t *testing.T) {
		part := newPart(t, db, testReference+"111", "Oil filter", 8290, 4)

		body := do(t, db, http.MethodGet, fmt.Sprintf("/parts/%d/edit", part.ID)).Body.String()

		// 8290 millimes must come back as 8.290, not 8.29 or 8290.
		if !strings.Contains(body, `value="8.290"`) {
			t.Error("the edit form does not show the price in dinars")
		}
	})

	t.Run("saves the changes", func(t *testing.T) {
		part := newPart(t, db, testReference+"222", "Oil filter", 8000, 4)

		got := post(t, db, fmt.Sprintf("/parts/%d", part.ID), url.Values{
			"reference": {testReference + "222"},
			"name":      {"Oil filter, long life"},
			"price":     {"9.750"},
			"quantity":  {"10"},
		})

		if got.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusSeeOther)
		}

		stored, err := db.Part(context.Background(), part.ID)
		if err != nil {
			t.Fatalf("Part() failed: %v", err)
		}
		if stored.PriceMillimes != 9750 || stored.QuantityOnHand != 10 {
			t.Errorf("stored = %+v, want the updated price and quantity", stored)
		}
	})

	t.Run("deletes the part", func(t *testing.T) {
		part, err := db.CreatePart(context.Background(), "TMP-999", "Temporary", 1000, 1)
		if err != nil {
			t.Fatalf("CreatePart() failed: %v", err)
		}

		if got := post(t, db, fmt.Sprintf("/parts/%d/delete", part.ID), url.Values{}); got.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, want %d", got.Code, http.StatusSeeOther)
		}

		if _, err := db.Part(context.Background(), part.ID); !errors.Is(err, store.ErrNotFound) {
			t.Errorf("Part() error = %v, want ErrNotFound", err)
		}
	})

	t.Run("answers 404 for a part that does not exist", func(t *testing.T) {
		if got := do(t, db, http.MethodGet, "/parts/999999999/edit"); got.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", got.Code, http.StatusNotFound)
		}
	})
}
