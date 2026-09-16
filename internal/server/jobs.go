package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Mariem-Abdennabi/car-service-management/internal/store"
	"github.com/Mariem-Abdennabi/car-service-management/views"
)

// handleJobs lists every service job, newest first.
func (s *Server) handleJobs(c *gin.Context) {
	jobs, err := s.store.Jobs(c.Request.Context())
	if err != nil {
		serverError(c, err)

		return
	}

	render(c, http.StatusOK, views.Jobs(s.assets, jobs))
}

// handleJob shows one job, the vehicle it is on, and that vehicle's owner.
func (s *Server) handleJob(c *gin.Context) {
	job, vehicle, customer, ok := s.jobFromPath(c)
	if !ok {
		return
	}

	render(c, http.StatusOK, views.JobDetail(s.assets, job, vehicle, customer))
}

// jobFromPath loads the job named by the :id path parameter, along with its
// vehicle and that vehicle's owner.
//
// All three come back together because every page about a job needs all three: to
// name the car, to say whose it is, and to link back. Following the chain from the
// job means the owner is always the real owner — the same reasoning as
// vehicleFromPath, one link further along.
func (s *Server) jobFromPath(c *gin.Context) (store.Job, store.Vehicle, store.Customer, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		s.notFound(c)

		return store.Job{}, store.Vehicle{}, store.Customer{}, false
	}

	ctx := c.Request.Context()

	job, err := s.store.Job(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		s.notFound(c)

		return store.Job{}, store.Vehicle{}, store.Customer{}, false
	}
	if err != nil {
		serverError(c, err)

		return store.Job{}, store.Vehicle{}, store.Customer{}, false
	}

	// A foreign key guarantees these two exist, so a missing row here is a broken
	// database rather than a bad URL — 500, not 404.
	vehicle, err := s.store.Vehicle(ctx, job.VehicleID)
	if err != nil {
		serverError(c, err)

		return store.Job{}, store.Vehicle{}, store.Customer{}, false
	}

	customer, err := s.store.Customer(ctx, vehicle.CustomerID)
	if err != nil {
		serverError(c, err)

		return store.Job{}, store.Vehicle{}, store.Customer{}, false
	}

	return job, vehicle, customer, true
}
