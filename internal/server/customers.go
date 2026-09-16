package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Mariem-Abdennabi/car-service-management/internal/store"
	"github.com/Mariem-Abdennabi/car-service-management/views"
)

// handleCustomers lists customers, optionally filtered by ?q=.
//
// It answers both a full page and a bare fragment. htmx sets the HX-Request
// header, and a fragment response is simply a component rendered without the
// layout around it — which is why searching needs no second route.
func (s *Server) handleCustomers(c *gin.Context) {
	search := strings.TrimSpace(c.Query("q"))

	customers, err := s.store.Customers(c.Request.Context(), search)
	if err != nil {
		serverError(c, err)

		return
	}

	if c.GetHeader("HX-Request") == "true" {
		render(c, http.StatusOK, views.CustomerList(customers))

		return
	}

	render(c, http.StatusOK, views.Customers(s.assets, customers, search))
}

// handleCustomer shows one customer and their vehicles.
func (s *Server) handleCustomer(c *gin.Context) {
	customer, ok := s.customerFromPath(c)
	if !ok {
		return
	}

	vehicles, err := s.store.VehiclesByCustomer(c.Request.Context(), customer.ID)
	if err != nil {
		serverError(c, err)

		return
	}

	render(c, http.StatusOK, views.CustomerDetail(s.assets, customer, vehicles, ""))
}

// handleNewCustomer shows the empty create form.
func (s *Server) handleNewCustomer(c *gin.Context) {
	render(c, http.StatusOK, views.CustomerNew(s.assets, views.CustomerForm{}))
}

// handleCreateCustomer validates the submission and inserts the customer.
func (s *Server) handleCreateCustomer(c *gin.Context) {
	// Trimmed, so a field of spaces counts as empty rather than being stored.
	form := views.CustomerForm{
		Name:  strings.TrimSpace(c.PostForm("name")),
		Phone: strings.TrimSpace(c.PostForm("phone")),
		City:  strings.TrimSpace(c.PostForm("city")),
	}

	if form.Errors = validateCustomer(form); len(form.Errors) > 0 {
		// The same component as the empty form, redrawn with what was typed. 422
		// rather than 400: the request was well formed, its contents were not.
		render(c, http.StatusUnprocessableEntity, views.CustomerNew(s.assets, form))

		return
	}

	customer, err := s.store.CreateCustomer(c.Request.Context(), form.Name, form.Phone, form.City)
	if err != nil {
		serverError(c, err)

		return
	}

	// 303 rather than 302, so the browser follows with GET. Without the redirect, a
	// refresh would re-submit the form and insert the customer twice.
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/customers/%d", customer.ID))
}

// validateCustomer returns one message per invalid field, keyed by field name.
//
// Validation lives here for now. It moves to a service layer if a rule ever needs
// the database — "this phone number is already registered", say.
func validateCustomer(form views.CustomerForm) map[string]string {
	errs := make(map[string]string)

	if msg := firstOf(required("Name", form.Name), tooLong("Name", form.Name, maxNameLength)); msg != "" {
		errs["name"] = msg
	}
	if msg := validatePhone(form.Phone); msg != "" {
		errs["phone"] = msg
	}
	if msg := firstOf(required("City", form.City), tooLong("City", form.City, maxCityLength)); msg != "" {
		errs["city"] = msg
	}

	return errs
}

// firstOf returns the first non-empty message, so checks can be listed in order of
// what the visitor should hear about first.
func firstOf(messages ...string) string {
	for _, msg := range messages {
		if msg != "" {
			return msg
		}
	}

	return ""
}

// handleEditCustomer shows the edit form with the customer's current details.
func (s *Server) handleEditCustomer(c *gin.Context) {
	customer, ok := s.customerFromPath(c)
	if !ok {
		return
	}

	render(c, http.StatusOK, views.CustomerEdit(s.assets, customer, views.CustomerForm{
		Name:  customer.Name,
		Phone: customer.Phone,
		City:  customer.City,
	}))
}

// handleUpdateCustomer validates the submission and saves the changes.
func (s *Server) handleUpdateCustomer(c *gin.Context) {
	customer, ok := s.customerFromPath(c)
	if !ok {
		return
	}

	form := views.CustomerForm{
		Name:  strings.TrimSpace(c.PostForm("name")),
		Phone: strings.TrimSpace(c.PostForm("phone")),
		City:  strings.TrimSpace(c.PostForm("city")),
	}

	if form.Errors = validateCustomer(form); len(form.Errors) > 0 {
		render(c, http.StatusUnprocessableEntity, views.CustomerEdit(s.assets, customer, form))

		return
	}

	updated, err := s.store.UpdateCustomer(c.Request.Context(), customer.ID, form.Name, form.Phone, form.City)
	if err != nil {
		serverError(c, err)

		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/customers/%d", updated.ID))
}

// handleDeleteCustomer removes a customer and returns to the list.
//
// A POST, not a DELETE: HTML forms only support GET and POST, so this works with
// JavaScript switched off. htmx can issue the same POST at step 3e.
func (s *Server) handleDeleteCustomer(c *gin.Context) {
	customer, ok := s.customerFromPath(c)
	if !ok {
		return
	}

	err := s.store.DeleteCustomer(c.Request.Context(), customer.ID)
	if errors.Is(err, store.ErrInUse) {
		// The vehicles table refused it. Explain that on the page rather than
		// answering 500, which is what an unexplained constraint violation becomes.
		s.renderCustomerWithNotice(c, customer, "This customer still has vehicles. Delete those first.")

		return
	}
	if err != nil {
		serverError(c, err)

		return
	}

	c.Redirect(http.StatusSeeOther, "/customers")
}

// customerFromPath loads the customer named by the :id path parameter.
//
// It answers 404 or 500 itself and reports false, so callers can return
// immediately — the three handlers above all need exactly this.
func (s *Server) customerFromPath(c *gin.Context) (store.Customer, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		s.notFound(c)

		return store.Customer{}, false
	}

	customer, err := s.store.Customer(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		s.notFound(c)

		return store.Customer{}, false
	}
	if err != nil {
		serverError(c, err)

		return store.Customer{}, false
	}

	return customer, true
}

// renderCustomerWithNotice redraws the detail page with a message, using 409 to say
// the request was understood but conflicts with the current state.
func (s *Server) renderCustomerWithNotice(c *gin.Context, customer store.Customer, notice string) {
	vehicles, err := s.store.VehiclesByCustomer(c.Request.Context(), customer.ID)
	if err != nil {
		serverError(c, err)

		return
	}

	render(c, http.StatusConflict, views.CustomerDetail(s.assets, customer, vehicles, notice))
}
