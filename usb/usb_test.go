package usb_test

import (
	"testing"

	. "github.com/kylelemons/gousb/usb"
	"github.com/kylelemons/gousb/usbid"
)

func TestNoop(t *testing.T) {
	c := NewContext()
	defer c.Close()
	c.Debug(3)
}

func TestEnum(t *testing.T) {
	c := NewContext()
	defer c.Close()
	c.Debug(3)

	devs, err := c.ListDevices(func(bus, addr int, desc *Descriptor) bool {
		t.Logf("%03d:%03d %+v", bus, addr, desc)
		if v, ok := usbid.Vendors[desc.Vendor]; ok {
			if p, ok := v.Devices[desc.Product]; ok {
				t.Logf(" - %s (%s) %s (%s)", v, desc.Vendor, p, desc.Product)
			} else {
				t.Logf(" - %s (%s) Unknown", v, desc.Vendor)
			}
		}
		return false
	})
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()
	if err != nil {
		t.Fatalf("list: %s", err)
	}
}
