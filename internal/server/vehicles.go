package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Mariem-Abdennabi/car-service-management/internal/store"
	"github.com/Mariem-Abdennabi/car-service-management/views"
)

// earliestVehicleYear is a sanity bound, not a rule about cars. It exists to catch
// a typed-in 219 or 20222, not to make a claim about motoring history.
const earliestVehicleYear = 1900

// handleNewVehicle shows the form for adding a vehicle to a customer.
func (s *Server) handleNewVehicle(c *gin.Context) {
	customer, ok := s.customerFromPath(c)
	if !ok {
		return
	}

	render(c, http.StatusOK, views.VehicleNew(s.assets, customer, views.VehicleForm{}))
}

// handleCreateVehicle validates the submission and adds the vehicle.
func (s *Server) handleCreateVehicle(c *gin.Context) {
	customer, ok := s.customerFromPath(c)
	if !ok {
		return
	}

	form := views.VehicleForm{
		Plate: strings.TrimSpace(c.PostForm("plate")),
		Make:  strings.TrimSpace(c.PostForm("make")),
		Model: strings.TrimSpace(c.PostForm("model")),
		Year:  strings.TrimSpace(c.PostForm("year")),
	}

	year, errs := validateVehicle(form)
	if form.Errors = errs; len(form.Errors) > 0 {
		render(c, http.StatusUnprocessableEntity, views.VehicleNew(s.assets, customer, form))

		return
	}

	_, err := s.store.CreateVehicle(c.Request.Context(), customer.ID, form.Plate, form.Make, form.Model, year)
	if err != nil {
		serverError(c, err)

		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/customers/%d", customer.ID))
}

// validateVehicle returns the parsed year and one message per invalid field.
//
// The year is returned rather than re-parsed by the caller, because parsing it is
// how it gets validated — doing that twice invites the two from drifting apart.
func validateVehicle(form views.VehicleForm) (int32, map[string]string) {
	errs := make(map[string]string)

	if msg := firstOf(required("Plate", form.Plate), tooLong("Plate", form.Plate, maxPlateLength)); msg != "" {
		errs["plate"] = msg
	}
	if msg := firstOf(required("Make", form.Make), tooLong("Make", form.Make, maxNameLength)); msg != "" {
		errs["make"] = msg
	}
	if msg := firstOf(required("Model", form.Model), tooLong("Model", form.Model, maxNameLength)); msg != "" {
		errs["model"] = msg
	}

	// Next year is allowed: new models are sold before the year they are named for.
	latest := time.Now().Year() + 1

	year, err := strconv.Atoi(form.Year)
	switch {
	case form.Year == "":
		errs["year"] = "Year is required."
	case err != nil:
		errs["year"] = "Year must be a number."
	case year < earliestVehicleYear || year > latest:
		errs["year"] = fmt.Sprintf("Year must be between %d and %d.", earliestVehicleYear, latest)
	}

	return int32(year), errs
}

// handleEditVehicle shows the edit form with the vehicle's current details.
func (s *Server) handleEditVehicle(c *gin.Context) {
	vehicle, customer, ok := s.vehicleFromPath(c)
	if !ok {
		return
	}

	render(c, http.StatusOK, views.VehicleEdit(s.assets, customer, vehicle, views.VehicleForm{
		Plate: vehicle.Plate,
		Make:  vehicle.Make,
		Model: vehicle.Model,
		Year:  strconv.Itoa(int(vehicle.Year)),
	}))
}

// handleUpdateVehicle validates the submission and saves the changes.
func (s *Server) handleUpdateVehicle(c *gin.Context) {
	vehicle, customer, ok := s.vehicleFromPath(c)
	if !ok {
		return
	}

	form := views.VehicleForm{
		Plate: strings.TrimSpace(c.PostForm("plate")),
		Make:  strings.TrimSpace(c.PostForm("make")),
		Model: strings.TrimSpace(c.PostForm("model")),
		Year:  strings.TrimSpace(c.PostForm("year")),
	}

	year, errs := validateVehicle(form)
	if form.Errors = errs; len(form.Errors) > 0 {
		render(c, http.StatusUnprocessableEntity, views.VehicleEdit(s.assets, customer, vehicle, form))

		return
	}

	if _, err := s.store.UpdateVehicle(c.Request.Context(), vehicle.ID, form.Plate, form.Make, form.Model, year); err != nil {
		serverError(c, err)

		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/customers/%d", customer.ID))
}

// handleDeleteVehicle removes a vehicle and returns to its customer.
func (s *Server) handleDeleteVehicle(c *gin.Context) {
	vehicle, customer, ok := s.vehicleFromPath(c)
	if !ok {
		return
	}

	err := s.store.DeleteVehicle(c.Request.Context(), vehicle.ID)
	if errors.Is(err, store.ErrInUse) {
		// The same shape as refusing to delete a customer with vehicles: the page
		// comes back with the reason rather than a 500.
		s.renderCustomerWithNotice(c, customer, "This vehicle has service jobs. Those have to go first.")

		return
	}
	if err != nil {
		serverError(c, err)

		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/customers/%d", customer.ID))
}

// vehicleFromPath loads the vehicle named by :id, and the customer it belongs to.
//
// The customer comes back too because every one of these handlers needs it — to
// link back, to name the page, and to redirect afterwards. Loading it here means
// the vehicle's owner is read from the vehicle rather than from the URL, so there
// is no way to act on one customer's vehicle while displaying another's name.
func (s *Server) vehicleFromPath(c *gin.Context) (store.Vehicle, store.Customer, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		s.notFound(c)

		return store.Vehicle{}, store.Customer{}, false
	}

	vehicle, err := s.store.Vehicle(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		s.notFound(c)

		return store.Vehicle{}, store.Customer{}, false
	}
	if err != nil {
		serverError(c, err)

		return store.Vehicle{}, store.Customer{}, false
	}

	customer, err := s.store.Customer(c.Request.Context(), vehicle.CustomerID)
	if err != nil {
		serverError(c, err)

		return store.Vehicle{}, store.Customer{}, false
	}

	return vehicle, customer, true
}
