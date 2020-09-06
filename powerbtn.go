package main

import (
	"time"

	evdev "github.com/gvalkov/golang-evdev"
	"github.com/stianeikeland/go-rpio"
)

type state int

const (
	stateIdle = iota
	stateDown
	stateShort
	stateLong
	stateCool
)

func (s state) String() string {
	return [...]string{"Idle", "Down", "Short", "Long", "Cool"}[s]
}

type fsm struct {
	current state
	timer   *time.Timer
	relay   rpio.Pin
}

func (m *fsm) enterState(newstate state) {
	if m.current == newstate {
		return
	}

	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}

	switch newstate {
	case stateDown:
		m.timer = time.AfterFunc(time.Second*2, func() {
			m.enterState(stateLong)
		})
	case stateShort:
		m.relay.High()
		m.timer = time.AfterFunc(time.Millisecond*250, func() {
			m.enterState(stateCool)
		})
	case stateLong:
		m.relay.High()
		m.timer = time.AfterFunc(time.Second*5, func() {
			m.enterState(stateCool)
		})
	case stateCool:
		m.relay.Low()
		m.timer = time.AfterFunc(time.Second*2, func() {
			m.enterState(stateIdle)
		})
	}

	m.current = newstate
}

func (m *fsm) onKeyEventState(s evdev.KeyEventState) {
	switch m.current {
	case stateIdle:
		if s == evdev.KeyDown {
			m.enterState(stateDown)
		}
	case stateDown:
		if s == evdev.KeyUp {
			m.enterState(stateShort)
		}
	}
}

// PowerButtonHandler handles presses to the power button to toggle a relay wired to the physical power button
func PowerButtonHandler(in <-chan KeyEvent) (<-chan KeyEvent, <-chan error, error) {

	err := rpio.Open()
	if err != nil {
		return nil, nil, err
	}

	var m fsm
	m.relay = rpio.Pin(16)
	m.relay.Output()
	m.relay.Low()

	out := make(chan KeyEvent)
	chErr := make(chan error)

	go func() {
		defer rpio.Close()
		defer close(out)
		defer close(chErr)

		var e KeyEvent
		for e = range in {
			switch e.Code {
			case evdev.KEY_POWER:
				m.onKeyEventState(e.State)
			default:
				// Send the event down the pipeline
				out <- e
			}
		}
	}()

	return out, chErr, nil
}
