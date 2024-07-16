package log

import "time"

type Clock interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

var Real Clock = realClock{}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}

func (realClock) Since(ts time.Time) time.Duration {
	return time.Since(ts)
}
