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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestLoaded(t *testing.T) {
	if got, min := len(Vendors), 1000; got < min {
		t.Errorf("%d vendors loaded, want at least %d", got, min)
	}
	if got, min := len(Classes), 10; got < min {
		t.Errorf("%d classes loaded, want at least %d", got, min)
	}
}

type handler struct{}

func (handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(testDBPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Open(%q): %v", testDBPath, err)
		return
	}
	defer f.Close()
	w.Header().Set("content-type", "text/plain")
	io.Copy(w, f)
}

func TestLoadFromURL(t *testing.T) {
	origV, origC := Vendors, Classes
	Vendors, Classes = nil, nil
	defer func() { Vendors, Classes = origV, origC }()

	s := httptest.NewServer(handler{})
	defer s.Close()

	err := LoadFromURL(s.URL)
	if err != nil {
		t.Fatalf("LoadFromURL(%q): got unexpected error: %v", s.URL, err)
	}
	if !reflect.DeepEqual(Vendors, testDBVendors) {
		t.Errorf("LoadFromURL Vendors:\ngot: %v\nwant: %v", Vendors, testDBVendors)
	}
}

func TestLoadFromURLError(t *testing.T) {
	origV, origC := Vendors, Classes
	defer func() { Vendors, Classes = origV, origC }()

	s := httptest.NewServer(http.NotFoundHandler())
	defer s.Close()

	ts := LastUpdate
	err := LoadFromURL(s.URL)
	if err == nil {
		t.Fatalf("LoadFromURL(%q): err is nil, want not nil", s.URL)
	}
	if LastUpdate != ts {
		t.Errorf("LastUpdate was unexpectedly updated when LoadFromURL failed")
	}
}
