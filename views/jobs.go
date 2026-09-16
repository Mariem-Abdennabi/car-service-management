package views

import "github.com/Mariem-Abdennabi/car-service-management/internal/store"

// JobStatusLabel turns a stored status into what a person reads: "in_progress"
// becomes "In progress".
//
// The map is here rather than in the store because it is a display concern — the
// database and the Go code agree on the stored value, and this is the only place
// that decides how it is worded.
func JobStatusLabel(status string) string {
	switch status {
	case store.JobReceived:
		return "Received"
	case store.JobInProgress:
		return "In progress"
	case store.JobCompleted:
		return "Completed"
	case store.JobCancelled:
		return "Cancelled"
	default:
		// A status the database allows but this switch has not been taught. Showing
		// the raw value is better than showing nothing at all.
		return status
	}
}

// jobStatusClasses is the badge's colour for a status. Cancelled and completed are
// deliberately quiet; in progress is the one worth spotting on a busy list.
func jobStatusClasses(status string) string {
	switch status {
	case store.JobInProgress:
		return "bg-brand-50 text-brand-700"
	case store.JobCompleted:
		return "bg-emerald-50 text-emerald-700"
	case store.JobCancelled:
		return "bg-slate-100 text-slate-500 line-through"
	default:
		return "bg-amber-50 text-amber-800"
	}
}
