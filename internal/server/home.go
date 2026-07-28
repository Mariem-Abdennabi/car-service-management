package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Mariem-Abdennabi/car-service-management/views"
)

// handleHome renders the landing page.
func (s *Server) handleHome(c *gin.Context) {
	render(c, http.StatusOK, views.Home(s.assets))
}
