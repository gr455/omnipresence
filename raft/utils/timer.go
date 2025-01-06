package utils

import (
	"log"
	"sync"
	"time"
)

type Timer struct {
	Interval              time.Duration
	Timer                 *time.Timer
	Enabled               bool
	Callback              func()
	CancelChannel         chan struct{}
	RoutineListeningMutex sync.Mutex
}

func NewTimer(intervalMillis uint16, callback func()) *Timer {
	timer := &Timer{}
	return timer.Initialize(intervalMillis, callback)
}

func (timer *Timer) Initialize(intervalMillis uint16, callback func()) *Timer {
	timer.Interval = time.Duration(intervalMillis) * time.Millisecond
	timer.Callback = callback
	timer.CancelChannel = make(chan struct{}, 0)
	timer.Timer = time.NewTimer(timer.Interval)
	timer.Timer.Stop()
	timer.Enabled = false

	return timer
}

func (timer *Timer) RestartIfEnabled() bool {
	if !timer.Enabled {
		return false
	}
	if unlocked := timer.RoutineListeningMutex.TryLock(); !unlocked {
		timer.CancelChannel <- struct{}{}
	} else {
		timer.RoutineListeningMutex.Unlock()
	}

	timer.Stop()
	timer.Timer.Reset(timer.Interval)

	go func() {
		// timer.RoutineListeningMutex.Lock()
		// defer timer.RoutineListeningMutex.Unlock()
		select {
		case <-timer.Timer.C:
			log.Printf("TIMER: %v", timer.Interval)
			timer.Callback()
		case <-timer.CancelChannel:
			break

		}
	}()

	return true
}

func (timer *Timer) Stop() {
	timer.Timer.Stop()
	// flush channel in case timer had written to it without goroutine consuming.
	select {
	case <-timer.Timer.C:
	default:
	}
}

func (timer *Timer) Disable() {
	timer.Stop()
	timer.Enabled = false
}

func (timer *Timer) Enable() {
	timer.Enabled = true
}
