package visualizer

import "time"

func GenFrame(start, now time.Time) time.Time {
	presets := []time.Duration{
		1 * time.Minute,
		5 * time.Minute,
		10 * time.Minute,
		30 * time.Minute,
		1 * time.Hour,
		2 * time.Hour,
		4 * time.Hour,
		8 * time.Hour,
		12 * time.Hour,
		24 * time.Hour,
	}

	for _, p := range presets {
		if now.Sub(start) < p {
			return start.Add(p)
		}
	}

	return start.Add(24 * time.Hour).Add(now.Sub(start).Truncate(24 * time.Hour))
}
