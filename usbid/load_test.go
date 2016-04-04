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

package usbid

import "testing"

func TestLoaded(t *testing.T) {
	if got, min := len(Vendors), 1000; got < min {
		t.Errorf("%d vendors loaded, want at least %d", got, min)
	}
	if got, min := len(Classes), 10; got < min {
		t.Errorf("%d classes loaded, want at least %d", got, min)
	}
}
