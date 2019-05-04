package main

import (
	"testing"
	"time"
)

func TestTimeToAge(t *testing.T) {
	tt := []struct {
		name     string
		diff     time.Duration
		expected string
	}{
		{
			"Less than a second",
			time.Duration(100 * time.Millisecond),
			"1s",
		},
		{
			"10 seconds",
			time.Duration(10 * time.Second),
			"10s",
		},
		{
			"60 seconds",
			time.Duration(60 * time.Second),
			"1m",
		},
		{
			"10 minutes",
			time.Duration(10 * time.Minute),
			"10m",
		},
		{
			"60 minutes",
			time.Duration(60 * time.Minute),
			"1h",
		},
		{
			"2 hours",
			time.Duration(125 * time.Minute),
			"2h",
		},
		{
			"24 hours",
			time.Duration(24 * time.Hour),
			"1d",
		},
		{
			"2 days",
			time.Duration(50 * time.Hour),
			"2d",
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			now := time.Now()
			result := timeToAge(now, now.Add(test.diff))

			if test.expected != result {
				t.Errorf("[%s] Expected: '%s', result: '%s'", test.name, test.expected, result)
			}
		})
	}
}
