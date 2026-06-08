package samplejob

import "time"

// Schedule decides when RunOnce should fire next. Replace with cron parsing
// (e.g. github.com/robfig/cron/v3) when you need richer triggers.
type Schedule struct {
	Interval time.Duration
}

// SleepUntilNextSlot returns how long to wait before the next run.
func (s Schedule) SleepUntilNextSlot(_ time.Time) time.Duration {
	if s.Interval <= 0 {
		return 5 * time.Minute
	}
	return s.Interval
}
