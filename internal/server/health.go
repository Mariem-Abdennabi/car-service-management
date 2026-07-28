package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleHealth reports that the process is alive and able to serve requests.
//
// This is the one endpoint that is not a feature. Its audience is machines: a
// container runtime deciding whether to restart this instance, a load balancer
// deciding whether to send it traffic, an uptime check deciding whether to page
// somebody. Keeping it cheap and dependency-free means it answers "is the
// process healthy?" and nothing else.
//
// At Milestone 2 it also pings the database, since a server that cannot reach
// its database is not usefully alive.
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
