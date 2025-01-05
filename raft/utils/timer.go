package utils

import (
	"log"
	"sync"
	"time"
)

type Timer struct {
	Interval              time.Duration
	Timer                 *time.Timer
	Callback              func()
	CancelChannel         chan struct{}
	RoutineListening      bool
	RoutineListeningMutex sync.Mutex
}

func NewTimer(intervalMillis uint16, callback func()) *Timer {
	timer := &Timer{}
	return timer.Initialize(intervalMillis, callback)
}

func (timer *Timer) Initialize(intervalMillis uint16, callback func()) *Timer {
	timer.Interval = time.Duration(intervalMillis) * time.Millisecond
	timer.Callback = callback
	timer.CancelChannel = make(chan struct{}, 1)
	timer.Timer = time.NewTimer(timer.Interval)

	return timer
}

func (timer *Timer) Restart() bool {
	log.Printf("TIMER RESTARTED")
	if unlocked := timer.RoutineListeningMutex.TryLock(); !unlocked {
		timer.CancelChannel <- struct{}{}
	} else {
		timer.RoutineListeningMutex.Unlock()
	}
	b := timer.Timer.Reset(timer.Interval)
	go func() {
		timer.RoutineListeningMutex.Lock()
		defer timer.RoutineListeningMutex.Unlock()
		select {
		case _ = <-timer.Timer.C:
			timer.Callback()
		case _ = <-timer.CancelChannel:
			break

		}
	}()

	timer.RoutineListening = true
	return b
}

func (timer *Timer) Stop() bool {
	return timer.Timer.Stop()
}
