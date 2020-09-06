# irhid

Small program that reads an event interface (most likely coming from an IR remote) and generate HID reports that are then fed to a HID gadget device.

This program was written because I wanted to control a HTPC with an IR remote, including powering off *and on* the machine. The solution is to use a raspberry pi zero configured with a gpio ir receiver, a configured hid gadget device for the usb port and a relay attached to gpio pin 16. The relay is in turn connected to the power switch of the pc and the usb port is connected to an always on usb port.

As the program is designed to capture the power button, it needs to be run like this:
```
systemd-inhibit --what=handle-power-key ./irhid /dev/input/event0 /dev/hidg0
```
