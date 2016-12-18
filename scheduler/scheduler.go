// Run given task at specified interval, aligned with date/time.
// Interval must be 1s or more, and should be dividable into the greater time unit without a remainder, e.g.
// 15min divides into 1 Hour(60min) without remainder and will execute the given task at 00:00, 00:15, 00:30, etc.

package scheduler

import (
	"time"
)

type scheduler struct {
	t        *time.Timer
	interval time.Duration
	task     func(timestamp time.Time)
}

func NewScheduler(interval time.Duration, task func(timestamp time.Time)) *scheduler {
	s := scheduler{interval: interval, task: task}
	s.init()
	go func(s *scheduler) {
		for {
			trigTime, ok := <-s.t.C
			if !ok {
				return
			}
			go s.task(trigTime)
			s.t.Reset(s.interval - time.Duration(trigTime.Nanosecond()))
		}
	}(&s)
	return &s
}

func (s *scheduler) NewInterval(interval time.Duration) {
	s.interval = interval
	s.init()
}

func (s *scheduler) Stop() {
	s.t.Stop()
}

func (s *scheduler) init() {
	now := time.Now()
	intervalSec := int64(s.interval / time.Second)
	mod := now.Unix() % intervalSec
	s.t = time.NewTimer((time.Duration(intervalSec-mod) * time.Second) - time.Duration(now.Nanosecond()))
}
