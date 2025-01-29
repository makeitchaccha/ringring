package visualizer

import (
	"image/color"
	"time"
)

type Tics struct {
	interval time.Duration
	format   string
	color    color.Color
}

func CalculateTics(d time.Duration) (Tics, Tics) {
	base := chooseBaseDuration(d)
	upgrade := upgradeDuration(base)

	return Tics{
			interval: base,
			format:   chooseFormat(base),
			color:    color.RGBA{200, 200, 200, 255},
		},
		Tics{
			interval: upgrade,
			format:   chooseFormat(upgrade),
			color:    color.RGBA{100, 100, 100, 255},
		}
}

func chooseFormat(d time.Duration) string {
	if d < 60*time.Second {
		return "15:04:05"
	}
	if d < 24*time.Hour {
		return "15:04"
	}

	return "01/02"
}

func chooseBaseDuration(d time.Duration) time.Duration {
	presets := []time.Duration{
		10 * time.Second,
		30 * time.Second,
		1 * time.Minute,
		5 * time.Minute,
		10 * time.Minute,
		30 * time.Minute,
		1 * time.Hour,
		4 * time.Hour,
		8 * time.Hour,
		12 * time.Hour,
		24 * time.Hour,
	}

	for _, p := range presets {
		if d < 12*p {
			return p
		}
	}

	return 24 * time.Hour
}

func upgradeDuration(d time.Duration) time.Duration {
	if d < 24*time.Hour {
		return 24 * time.Hour
	}
	if d < 7*24*time.Hour {
		return 7 * 24 * time.Hour // 1 week (?)
	}

	return 30 * 24 * time.Hour // (??) would the call last for 30 days?
}
