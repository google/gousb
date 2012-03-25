package usb

import (
	"fmt"
)

type BCD uint16

const (
	USB_2_0 BCD = 0x0200
	USB_1_1 BCD = 0x0110
	USB_1_0 BCD = 0x0100
)

func (d BCD) Int() (i int) {
	ten := 1
	for o := uint(0); o < 4; o++ {
		n := ((0xF << (o*4)) & d) >> (o*4)
		i += int(n) * ten
		ten *= 10
	}
	return
}

func (d BCD) String() string {
	return fmt.Sprintf("%02x.%02x", int(d >> 8), int(d & 0xFF))
}

type ID uint16

func (id ID) String() string {
	return fmt.Sprintf("%04x", int(id))
}
