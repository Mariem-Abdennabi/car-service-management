package server

import (
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
)

// render writes a templ component as the response body.
//
// Every handler that returns HTML goes through here, so the Content-Type is set
// in exactly one place.
//
// Note what happens on failure. The header and status are committed as soon as
// the first byte of the body is written, so a component that fails half-way
// cannot be turned into a clean 500 — part of a page has already reached the
// browser. All that is left is to record the error, which c.Error attaches to the
// request for the middleware to report. Rendering a component is unlikely to fail
// in practice (it is writing strings to a socket), but silently discarding the
// error would hide a genuinely broken page.
func render(c *gin.Context, status int, component templ.Component) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(status)

	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		_ = c.Error(err)
	}
}
