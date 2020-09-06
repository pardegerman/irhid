// +build linux

package main

import (
	"fmt"
	"os"
)

const (
	usage                  = "usage: irhid [event device] [hid gadget device]"
	defaultEventDevice     = "/dev/input/event0"
	defaultHIDGadgetDevice = "/dev/hidg0"
)

func main() {
	devicePath := defaultEventDevice
	gadgetPath := defaultHIDGadgetDevice

	switch len(os.Args) {
	case 2:
		devicePath = os.Args[1]
	case 3:
		devicePath = os.Args[1]
		gadgetPath = os.Args[2]
	default:
		fmt.Printf(usage + "\n")
		os.Exit(1)
	}

	var err error

	// Start event source
	chEvents, chErrEv, err := StartReading(devicePath)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	// Sort out power button events
	chEventsToSend, chErrPwr, err := PowerButtonHandler(chEvents)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	// Transform events to hid reports
	chReports, chErrReport := ParseEvents(chEventsToSend)

	// Start HID gadget sink
	chErrHID, err := StartHidGadget(gadgetPath, chReports)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	go func() {
		for {
			select {
			case err = <-chErrEv:
				fmt.Printf("EV: %s\n", err)
			case err = <-chErrPwr:
				fmt.Printf("PWRBTN: %s\n", err)
			case err = <-chErrReport:
				fmt.Printf("REPORT: %s\n", err)
			case err = <-chErrHID:
				fmt.Printf("HID: %s\n", err)
			default:
			}
		}
	}()

	quit := make(chan struct{})
	<-quit
}
