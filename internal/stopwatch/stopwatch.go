package readunread

import (
	"time"
)

// Stopwatch struct with start time and offset
type Stopwatch struct {
	start time.Time
}

func New() Stopwatch {
	return Stopwatch{
		start: time.Now(), // Initialize start time with offset
	}
}

// Start resets the stopwatch and starts timing
func (s *Stopwatch) Start() {
	s.start = time.Now()
}

func (s *Stopwatch) Elapsed() time.Duration {
	return time.Since(s.start)
}
