package utils

import (
	"log"
	"time"
)

type Timer struct {
	Interval      time.Duration
	Timer         *time.Timer
	Callback      func()
	CancelChannel chan struct{}
}

func NewTimer(intervalMillis uint16, callback func()) *Timer {
	timer := &Timer{}
	log.Printf("HERE")
	return timer.Initialize(intervalMillis, callback)
}

func (timer *Timer) Initialize(intervalMillis uint16, callback func()) *Timer {
	timer.Interval = time.Duration(intervalMillis) * time.Millisecond
	timer.Callback = callback

	timer.Timer = time.NewTimer(timer.Interval)

	return timer
}

func (timer *Timer) Restart() bool {
	timer.CancelChannel <- struct{}{}
	b := timer.Timer.Reset(timer.Interval)
	go func() {
		select {
		case _ = <-timer.Timer.C:
			timer.Callback()
		case _ = <-timer.CancelChannel:
			break

		}
	}()

	return b
}

func (timer *Timer) Stop() bool {
	return timer.Timer.Stop()
}
