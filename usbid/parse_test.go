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

package usbid

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func mustRead(fname string) string {
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestParse(t *testing.T) {
	vendors, classes, err := ParseIDs(strings.NewReader(mustRead(testDBPath)))
	if err != nil {
		t.Fatalf("ParseIDs(%q): %s", testDBPath, err)
	}
	if got, want := vendors, testDBVendors; !reflect.DeepEqual(got, want) {
		t.Errorf("Vendors from %q", testDBPath)
		t.Errorf(" - got:  %+v", got)
		t.Errorf(" - want: %+v", want)
	}
	if got, want := classes, testDBClasses; !reflect.DeepEqual(got, want) {
		t.Errorf("Classes from %q", testDBPath)
		t.Errorf(" - got:  %+v", got)
		t.Errorf(" - want: %+v", want)
	}
}
