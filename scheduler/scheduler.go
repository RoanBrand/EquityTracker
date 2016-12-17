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
			s.task(trigTime)
			s.t.Reset(s.interval)
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
	if mod == 0 {
		s.task(now)
		s.t = time.NewTimer(s.interval)
	} else {
		s.t = time.NewTimer(time.Duration(intervalSec-mod) * time.Second)
	}
}
