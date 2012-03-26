package usb_test

import (
	"testing"

	. "github.com/kylelemons/gousb/usb"
	"github.com/kylelemons/gousb/usbid"
)

func TestNoop(t *testing.T) {
	c := NewContext()
	defer c.Close()
	c.Debug(0)
}

func TestEnum(t *testing.T) {
	c := NewContext()
	defer c.Close()
	c.Debug(0)

	cnt := 0
	devs, err := c.ListDevices(func(bus, addr int, desc *Descriptor) bool {
		cnt++
		return true
	})
	defer func() {
		for _, d := range devs {
			d.Close()
		}
	}()
	if err != nil {
		t.Fatalf("list: %s", err)
	}

	if got, want := len(devs), cnt; got != want {
		t.Errorf("len(devs) = %d, want %d", got, want)
	}

	for _, dev := range devs {
		desc, err := dev.Descriptor()
		if err != nil {
			t.Errorf("desc: %s", err)
			continue
		}
		bus := dev.BusNumber()
		addr := dev.Address()

		t.Logf("%03d:%03d %+v", bus, addr, desc)
		if v, ok := usbid.Vendors[desc.Vendor]; ok {
			if p, ok := v.Devices[desc.Product]; ok {
				t.Logf(" - %s (%s) %s (%s)", v, desc.Vendor, p, desc.Product)
			} else {
				t.Logf(" - %s (%s) Unknown", v, desc.Vendor)
			}
		}

		cfgs, err := dev.Configurations()
		defer func() {
			for _, cfg := range cfgs {
				cfg.Close()
			}
		}()
		if err != nil {
			t.Errorf("   - configs: %s", err)
			continue
		}

		for _, cfg := range cfgs {
			t.Logf("   - %#v", cfg)
			for _, alt := range cfg.Interfaces {
				for _, iface := range alt {
					t.Logf("     - %#v", iface)
					for _, end := range iface.Endpoints {
						t.Logf("       - %#v", end)
					}
				}
				t.Logf("     -----")
			}
		}
	}
}
