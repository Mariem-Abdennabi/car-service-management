package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Mariem-Abdennabi/car-service-management/internal/store"
	"github.com/Mariem-Abdennabi/car-service-management/views"
)

// handleParts lists the spare parts catalogue.
func (s *Server) handleParts(c *gin.Context) {
	parts, err := s.store.Parts(c.Request.Context())
	if err != nil {
		serverError(c, err)

		return
	}

	render(c, http.StatusOK, views.Parts(s.assets, parts))
}

// handleNewPart shows the empty create form.
func (s *Server) handleNewPart(c *gin.Context) {
	render(c, http.StatusOK, views.PartNew(s.assets, views.PartForm{Quantity: "0"}))
}

// handleCreatePart validates the submission and adds the part.
func (s *Server) handleCreatePart(c *gin.Context) {
	form := partFormFrom(c)

	price, quantity, errs := validatePart(form)
	if form.Errors = errs; len(form.Errors) > 0 {
		render(c, http.StatusUnprocessableEntity, views.PartNew(s.assets, form))

		return
	}

	_, err := s.store.CreatePart(c.Request.Context(), form.Reference, form.Name, price, quantity)
	if errors.Is(err, store.ErrDuplicate) {
		// A field error, not a 500: the visitor can fix this by changing one box.
		form.Errors = map[string]string{"reference": "That reference is already used by another part."}
		render(c, http.StatusUnprocessableEntity, views.PartNew(s.assets, form))

		return
	}
	if err != nil {
		serverError(c, err)

		return
	}

	c.Redirect(http.StatusSeeOther, "/parts")
}

// handleEditPart shows the edit form with the part's current details.
func (s *Server) handleEditPart(c *gin.Context) {
	part, ok := s.partFromPath(c)
	if !ok {
		return
	}

	render(c, http.StatusOK, views.PartEdit(s.assets, part, views.PartForm{
		Reference: part.Reference,
		Name:      part.Name,
		// Shown in the same form it is typed in, so the round trip is lossless.
		Price:    strings.TrimSuffix(views.Millimes(part.PriceMillimes), " TND"),
		Quantity: strconv.Itoa(int(part.QuantityOnHand)),
	}))
}

// handleUpdatePart validates the submission and saves the changes.
func (s *Server) handleUpdatePart(c *gin.Context) {
	part, ok := s.partFromPath(c)
	if !ok {
		return
	}

	form := partFormFrom(c)

	price, quantity, errs := validatePart(form)
	if form.Errors = errs; len(form.Errors) > 0 {
		render(c, http.StatusUnprocessableEntity, views.PartEdit(s.assets, part, form))

		return
	}

	_, err := s.store.UpdatePart(c.Request.Context(), part.ID, form.Reference, form.Name, price, quantity)
	if errors.Is(err, store.ErrDuplicate) {
		form.Errors = map[string]string{"reference": "That reference is already used by another part."}
		render(c, http.StatusUnprocessableEntity, views.PartEdit(s.assets, part, form))

		return
	}
	if err != nil {
		serverError(c, err)

		return
	}

	c.Redirect(http.StatusSeeOther, "/parts")
}

// handleDeletePart removes a part.
func (s *Server) handleDeletePart(c *gin.Context) {
	part, ok := s.partFromPath(c)
	if !ok {
		return
	}

	if err := s.store.DeletePart(c.Request.Context(), part.ID); err != nil {
		serverError(c, err)

		return
	}

	c.Redirect(http.StatusSeeOther, "/parts")
}

// partFromPath loads the part named by the :id path parameter.
func (s *Server) partFromPath(c *gin.Context) (store.Part, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		s.notFound(c)

		return store.Part{}, false
	}

	part, err := s.store.Part(c.Request.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		s.notFound(c)

		return store.Part{}, false
	}
	if err != nil {
		serverError(c, err)

		return store.Part{}, false
	}

	return part, true
}

// partFormFrom reads the submitted form, trimmed.
func partFormFrom(c *gin.Context) views.PartForm {
	return views.PartForm{
		Reference: strings.TrimSpace(c.PostForm("reference")),
		Name:      strings.TrimSpace(c.PostForm("name")),
		Price:     strings.TrimSpace(c.PostForm("price")),
		Quantity:  strings.TrimSpace(c.PostForm("quantity")),
	}
}

// validatePart returns the price in millimes, the quantity, and one message per
// invalid field.
func validatePart(form views.PartForm) (price, quantity int32, errs map[string]string) {
	errs = make(map[string]string)

	if msg := firstOf(required("Reference", form.Reference), tooLong("Reference", form.Reference, maxReferenceLength)); msg != "" {
		errs["reference"] = msg
	}
	if msg := firstOf(required("Name", form.Name), tooLong("Name", form.Name, maxNameLength)); msg != "" {
		errs["name"] = msg
	}

	switch parsed, err := parseMillimes(form.Price); {
	case form.Price == "":
		errs["price"] = "Price is required."
	case err != nil:
		errs["price"] = "Price must be an amount in dinars, like 42.500."
	default:
		price = parsed
	}

	switch parsed, err := strconv.Atoi(form.Quantity); {
	case form.Quantity == "":
		errs["quantity"] = "Quantity is required."
	case err != nil:
		errs["quantity"] = "Quantity must be a whole number."
	case parsed < 0:
		errs["quantity"] = "Quantity cannot be negative."
	case parsed > maxQuantity:
		errs["quantity"] = "Quantity looks like a mistake."
	default:
		quantity = int32(parsed)
	}

	return price, quantity, errs
}
