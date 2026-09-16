package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"

	"github.com/Mariem-Abdennabi/car-service-management/views"
)

// render writes a templ component as the response body, so the Content-Type is set
// in one place.
//
// Status and headers commit with the first byte of the body, so a component that
// fails half-way cannot become a clean 500 — part of the page has already been
// sent. All that is left is to record the error, which c.Error attaches to the
// request for the middleware to report.
func render(c *gin.Context, status int, component templ.Component) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(status)

	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		_ = c.Error(err)
	}
}

// notFound renders the 404 page.
func (s *Server) notFound(c *gin.Context) {
	render(c, http.StatusNotFound, views.NotFound(s.assets))
}

// serverError records the error and answers 500 in plain text.
//
// Plain text rather than a rendered page: this is the path where something is
// already broken, and rendering a template could fail again. The error itself is
// never sent to the browser.
func serverError(c *gin.Context, err error) {
	_ = c.Error(err)
	c.String(http.StatusInternalServerError, "internal server error")
}
