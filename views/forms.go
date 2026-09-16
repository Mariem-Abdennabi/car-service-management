package views

// CustomerForm is what the create form needs to draw itself: the values the
// visitor typed, and any errors keyed by field name.
//
// It exists so a rejected submission can be redrawn with the values still in it.
// Without it, a validation failure means an empty form and typing everything again.
type CustomerForm struct {
	Name   string
	Phone  string
	City   string
	Errors map[string]string
}

// VehicleForm is the create form's state.
//
// Year is a string, not an int: a rejected form has to redraw exactly what was
// typed, and "not a year" is not an int. Parsing belongs to validation, not to
// reading the request.
type VehicleForm struct {
	Plate  string
	Make   string
	Model  string
	Year   string
	Errors map[string]string
}

// PartForm is the part form's state. Price is the text the visitor typed, in
// dinars — the store keeps millimes, and converting is part of validation.
type PartForm struct {
	Reference string
	Name      string
	Price     string
	Quantity  string
	Errors    map[string]string
}
