// Copyright 2013 Google Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package usb_test

import (
	"bytes"
	"log"
	"os"
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

	logDevice := func(t *testing.T, desc *Descriptor) {
		t.Logf("%03d.%03d %s", desc.Bus, desc.Address, usbid.Describe(desc))
		t.Logf("- Protocol: %s", usbid.Classify(desc))

		for _, cfg := range desc.Configs {
			t.Logf("- %s:", cfg)
			for _, alt := range cfg.Interfaces {
				t.Logf("  --------------")
				for _, iface := range alt.Setups {
					t.Logf("  - %s", iface)
					t.Logf("    - %s", usbid.Classify(iface))
					for _, end := range iface.Endpoints {
						t.Logf("    - %s (packet size: %d bytes)", end, end.MaxPacketSize)
					}
				}
			}
			t.Logf("  --------------")
		}
	}

	descs := []*Descriptor{}
	devs, err := c.ListDevices(func(desc *Descriptor) bool {
		logDevice(t, desc)
		descs = append(descs, desc)
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

	if got, want := len(devs), len(descs); got != want {
		t.Fatalf("len(devs) = %d, want %d", got, want)
	}

	for i := range devs {
		if got, want := devs[i].Descriptor, descs[i]; got != want {
			t.Errorf("dev[%d].Descriptor = %p, want %p", i, got, want)
		}
	}
}

func TestMultipleContexts(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	for i := 0; i < 2; i++ {
		ctx := NewContext()
		_, err := ctx.ListDevices(func(desc *Descriptor) bool {
			return false
		})
		if err != nil {
			t.Fatal(err)
		}
		ctx.Close()
	}
	log.SetOutput(os.Stderr)
	if buf.Len() > 0 {
		t.Errorf("Non zero output to log, while testing:  %s", buf.String())
	}
}
