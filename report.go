package main

import (
	"errors"
	"fmt"

	evdev "github.com/gvalkov/golang-evdev"
)

// Report represents all data that is required to build up a HID report
type Report struct {
	Modifiers byte
	KeyCodes  [6]byte
}

// Bytes returns a byte representation of the report
func (r *Report) Bytes() []byte {
	buf := []byte{0}
	buf[0] = r.Modifiers
	copy(buf[2:7], r.KeyCodes[:])
	return buf
}

var translate = []byte{
	3, 41, 30, 31, 32, 33, 34, 35, 36, 37, 38,
	39, 45, 46, 42, 43, 20, 26, 8, 21, 23, 28, 24, 12, 18, 19,
	47, 48, 40, 224, 4, 22, 7, 9, 10, 11, 13, 14, 15, 51, 52,
	53, 225, 50, 29, 27, 6, 25, 5, 17, 16, 54, 55, 56, 229, 85,
	226, 44, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 83,
	71, 95, 96, 97, 86, 92, 93, 94, 87, 89, 90, 91, 98, 99, 0,
	148, 100, 68, 69, 135, 146, 147, 138, 136, 139, 140, 88, 228,
	84, 70, 230, 0, 74, 82, 75, 80, 79, 77, 81, 78, 73, 76, 0,
	239, 238, 237, 102, 103, 0, 72, 0, 133, 144, 145, 137, 227,
	231, 101, 243, 121, 118, 122, 119, 124, 116, 125, 244, 123,
	117, 0, 251, 0, 248, 0, 0, 0, 0, 0, 0, 0, 240, 0,
	249, 0, 0, 0, 0, 0, 241, 242, 0, 236, 0, 235, 232, 234,
	233, 0, 0, 0, 0, 0, 0, 250, 0, 0, 247, 245, 246, 182,
	183, 0, 0, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113,
	114}

func (r *Report) isFull() bool {
	for _, c := range r.KeyCodes {
		if c == 0 {
			return false
		}
	}

	return true
}

func (r *Report) addKeyEvent(e KeyEvent) error {
	if e.Code >= len(translate) {
		return fmt.Errorf("key %s does not correspond to valid HID keycode, only keys in page 0x07 are supported", e)
	}
	// Find the first empty slot
	for idx, c := range r.KeyCodes {
		if c == 0 {
			r.KeyCodes[idx] = translate[e.Code]
			return nil
		}
	}

	return errors.New("trying to add event to an already full report")
}

func (r *Report) zero() {
	r.Modifiers = 0
	r.KeyCodes = [6]byte{0}
}

// ParseEvents to HID reports
func ParseEvents(in <-chan KeyEvent) (<-chan Report, <-chan error) {
	out := make(chan Report)
	chErr := make(chan error)

	go func() {
		defer close(out)
		defer close(chErr)

		var e KeyEvent
		var r Report
		var err error

		for e = range in {
			switch e.State {
			case evdev.KeyDown:
				err = r.addKeyEvent(e)
				if err != nil {
					chErr <- err
				} else {
					out <- r
				}
				r.zero()

			case evdev.KeyUp:
				// Send a zero report to indicate that no keys are pressed
				r.zero()
				out <- r
			}
		}
	}()

	return out, chErr
}
