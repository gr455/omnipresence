package utils

import (
	"log"
	"time"
)

type Timer struct {
	Interval time.Duration
	Ticker   *time.Ticker
	Enabled  bool
	Callback func()
	Debug    bool
}

func NewTimer(intervalMillis uint16, callback func(), debug bool) *Timer {
	timer := &Timer{}
	return timer.Initialize(intervalMillis, callback)
}

func (timer *Timer) Initialize(intervalMillis uint16, callback func()) *Timer {
	timer.Interval = time.Duration(intervalMillis) * time.Millisecond
	timer.Callback = callback
	timer.Ticker = time.NewTicker(timer.Interval)
	timer.Ticker.Stop()
	timer.Enabled = false

	go func() {
		for range timer.Ticker.C {
			timer.debug("TICK")
			timer.Callback()
		}
	}()

	return timer
}

func (timer *Timer) RestartIfEnabled() bool {
	if !timer.Enabled {
		return false
	}

	timer.debug("RESTARTED")

	timer.Ticker.Stop()
	timer.Ticker.Reset(timer.Interval)
	return true
}

func (timer *Timer) Stop() {
	timer.Ticker.Stop()
}

func (timer *Timer) StopAndDisable() {
	timer.Stop()
	timer.Enabled = false
}

func (timer *Timer) Enable() {
	timer.Enabled = true
}

func (timer *Timer) debug(msg string) {
	if timer.Debug {
		log.Printf("\n\n\n\n %s\n\n\n", msg)
	}
}
