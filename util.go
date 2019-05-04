package main

import (
	"fmt"
	"time"
)

func timeToAge(start, end time.Time) string {
	diff := end.Sub(start)

	hours := diff.Truncate(time.Hour).Hours()
	if hours >= 24 {
		return fmt.Sprintf("%vd", int(hours/24))
	}
	if hours >= 1 && hours < 24 {
		return fmt.Sprintf("%vh", hours)
	}
	minutes := diff.Truncate(time.Minute).Minutes()
	if minutes >= 1 && minutes < 60 {
		return fmt.Sprintf("%vm", minutes)
	}
	seconds := diff.Truncate(time.Second).Seconds()
	if seconds > 1 && seconds < 60 {
		return fmt.Sprintf("%vs", seconds)
	}
	if seconds <= 1 {
		return "1s"
	}
	return ""
}
