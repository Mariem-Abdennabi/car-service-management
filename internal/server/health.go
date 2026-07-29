package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleHealth reports whether this instance can serve requests.
//
// Its audience is machines: a container runtime deciding whether to restart the
// instance, a load balancer deciding whether to send it traffic, an uptime check
// deciding whether to page somebody.
//
// It pings the database, because a server that cannot reach its database is not
// usefully alive. The distinction between the two failures matters to whoever is
// reading it: no response at all means the process is gone, while 503 means the
// process is up but not ready for traffic.
func (s *Server) handleHealth(c *gin.Context) {
	// The request's context, so a client that gives up does not leave the ping
	// waiting on a database that may itself be the problem.
	if err := s.store.Ping(c.Request.Context()); err != nil {
		// The error text is deliberately not returned: this endpoint is public
		// enough that it should not leak connection strings or hostnames.
		_ = c.Error(err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
