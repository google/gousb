// Copyright 2013 Google Inc.  All rights reserved.
// Copyright 2016 the gousb Authors.  All rights reserved.
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

package gousb

import (
	"testing"
)

func TestBCD(t *testing.T) {
	t.Parallel()
	tests := []struct {
		major, minor uint8
		bcd          BCD
		str          string
	}{
		{1, 1, 0x0101, "1.01"},
		{12, 34, 0x1234, "12.34"},
	}

	for _, test := range tests {
		bcd := Version(test.major, test.minor)
		if bcd != test.bcd {
			t.Errorf("Version(%d, %d): got BCD %04x, want %04x", test.major, test.minor, uint16(bcd), uint16(test.bcd))
			continue
		}
		if got, want := bcd.String(), test.str; got != want {
			t.Errorf("String(%04x) = %q, want %q", uint16(test.bcd), got, want)
		}
	}
}
