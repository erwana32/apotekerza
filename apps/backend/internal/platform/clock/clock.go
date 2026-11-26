// Package clock provides an injectable clock for deterministic tests.
package clock

import "time"

// Clock returns the current time in UTC.
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

// System returns a clock backed by time.Now in UTC.
func System() Clock { return systemClock{} }

type fixedClock struct{ moment time.Time }

func (f fixedClock) Now() time.Time { return f.moment }

// Fixed returns a clock that always reports moment in UTC.
func Fixed(moment time.Time) Clock { return fixedClock{moment: moment.UTC()} }
