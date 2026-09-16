package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Mariem-Abdennabi/car-service-management/views"
)

// handleHome renders the dashboard.
func (s *Server) handleHome(c *gin.Context) {
	ctx := c.Request.Context()

	summary, err := s.store.Summary(ctx)
	if err != nil {
		serverError(c, err)

		return
	}

	recent, err := s.store.RecentCustomers(ctx)
	if err != nil {
		serverError(c, err)

		return
	}

	outOfStock, err := s.store.PartsOutOfStock(ctx)
	if err != nil {
		serverError(c, err)

		return
	}

	render(c, http.StatusOK, views.Home(s.assets, summary, recent, outOfStock))
}
