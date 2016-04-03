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
	"testing"

	. "github.com/kylelemons/gousb/usb"
)

func TestGetStringDescriptorAscii(t *testing.T) {
	c := NewContext()
	defer c.Close()
	c.Debug(0)

	devices, err := c.ListDevices(func(desc *Descriptor) bool {
		return true
	})
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	for i, d := range devices {

		str, err := d.GetStringDescriptor(1)
		if err != nil {
			t.Fatalf("%s", err.Error())
		}

		str2, err := d.GetStringDescriptor(2)
		if err != nil {
			t.Fatalf("%s", err.Error())
		}

		t.Logf("%d: %s %s\n", i, str, str2)
		d.Close()
	}
}
