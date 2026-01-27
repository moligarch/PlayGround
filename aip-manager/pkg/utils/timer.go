// pkg/utils/timer.go
package utils

import (
	"fmt"
	"time"
)

// StartTimer prints a message and returns the current time.
func StartTimer(message string) time.Time {
	fmt.Printf(" > %s\n", message)
	return time.Now()
}

// EndTimer takes a start time and prints the elapsed duration.
func EndTimer(start time.Time) {
	duration := time.Since(start)
	fmt.Printf("\n   ...✅ Done in %s\n", duration.Round(time.Millisecond))
}
