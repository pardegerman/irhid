package main

import (
	"fmt"

	evdev "github.com/gvalkov/golang-evdev"
)

// KeyEvent represents an event where a key is pressed or released
type KeyEvent struct {
	Code  int
	State evdev.KeyEventState
}

func (k KeyEvent) String() string {
	return fmt.Sprintf("%s (%d)", evdev.KEY[k.Code], k.Code)
}

// StartReading opens a device and starts a goroutine that consume events, and generates press, long press and release events on the respective channels
func StartReading(devicepath string) (<-chan KeyEvent, <-chan error, error) {
	dev, err := evdev.Open(devicepath)
	if err != nil {
		return nil, nil, err
	}

	out := make(chan KeyEvent)
	chError := make(chan error)

	go func() {
		defer close(out)
		defer close(chError)

		var e KeyEvent
		for {
			events, err := dev.Read()
			if err != nil {
				chError <- err
				return
			}
			for _, ev := range events {
				switch ev.Type {
				case evdev.EV_KEY:
					e.Code = int(ev.Code)
					e.State = evdev.KeyEventState(ev.Value)
					out <- e
				}
			}
		}
	}()

	return out, chError, nil
}
