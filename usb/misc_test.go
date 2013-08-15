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

package usb

import (
	"testing"
)

func TestBCD(t *testing.T) {
	tests := []struct {
		BCD
		Int int
		Str string
	}{
		{0x1234, 1234, "12.34"},
	}

	for _, test := range tests {
		if got, want := test.BCD.Int(), test.Int; got != want {
			t.Errorf("Int(%x) = %d, want %d", test.BCD, got, want)
		}
		if got, want := test.BCD.String(), test.Str; got != want {
			t.Errorf("String(%x) = %q, want %q", test.BCD, got, want)
		}
	}
}
