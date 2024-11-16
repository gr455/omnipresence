package utils

import (
	"time"
)

type Timer struct {
	Interval time.Duration
	Timer    time.Timer
	Callback func()
}

func NewTimer(intervalMillis uint16, callback func()) *Timer {
	timer := &Timer{}
	return timer.Initialize(intervalMillis, callback)
}

func (timer *Timer) Initialize(intervalMillis uint16, callback func()) *Timer {
	timer.Interval = time.Duration(intervalMillis) * time.Millisecond
	timer.Callback = callback

	return timer
}

func (timer *Timer) Restart() bool {
	return timer.Timer.Reset(timer.Interval)
}

func (timer *Timer) Stop() bool {
	return timer.Timer.Stop()
}
