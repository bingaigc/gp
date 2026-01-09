package entity

import "time"

// Now returns the current time (helper for easier mocking in tests)
func Now() time.Time {
	return time.Now()
}
