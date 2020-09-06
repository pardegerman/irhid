package main

import (
	"fmt"
	"os"
)

// StartHidGadget opens a the hid gadget device and starts a goroutine that generate reports and sends them to the device
func StartHidGadget(path string, chReports <-chan Report) (<-chan error, error) {
	dev, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	chErr := make(chan error)

	go func() {
		defer close(chErr)
		defer dev.Close()

		var r Report

		for r = range chReports {
			buf := []byte{r.Modifiers, 0, r.KeyCodes[0], r.KeyCodes[1], r.KeyCodes[2], r.KeyCodes[3], r.KeyCodes[4], r.KeyCodes[5]}
			n, _ := dev.Write(buf)
			if n != len(buf) {
				err = fmt.Errorf("hid: write failed, expected to write %d bytes, wrote %d bytes", len(buf), n)
				chErr <- err
			}
		}
	}()

	return chErr, nil
}
