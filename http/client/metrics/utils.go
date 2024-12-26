package metrics

import (
	"time"
)

const (
	PKGLabelName = "pkg"
	Namespace    = ""
)

// MeasureDuration calculates the time elapsed since the start time and returns it in seconds.
func MeasureDuration(start time.Time) float64 {
	return time.Since(start).Seconds()
}
