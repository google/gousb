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

package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/google/gousb/usbid"
)

var (
	remote   = flag.String("url", usbid.LinuxUsbDotOrg, "URL from which to download new vendor data")
	dataFile = flag.String("template", "load_data.go.tpl", "Template filename")
	outFile  = flag.String("o", "load_data.go", "Output filename")
)

func main() {
	flag.Parse()

	log.Printf("Fetching %q...", *remote)
	resp, err := http.Get(*remote)
	if err != nil {
		log.Fatalf("failed to download from %q: %s", *remote, err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read %q: %s", *remote, err)
	}

	ids, cls, err := usbid.ParseIDs(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("failed to parse %q: %s", *remote, err)
	}

	log.Printf("Successfully fetched %q:", *remote)
	log.Printf("  Loaded %d Vendor IDs", len(ids))
	log.Printf("  Loaded %d Class IDs", len(cls))

	rawTemplate, err := ioutil.ReadFile(*dataFile)
	if err != nil {
		log.Fatalf("failed to read template %q: %s", *dataFile, err)
	}

	template, err := template.New("").Parse(string(rawTemplate))
	if err != nil {
		log.Fatalf("failed to parse template %q: %s", *dataFile, err)
	}

	out, err := os.Create(*outFile)
	if err != nil {
		log.Fatalf("failed to open output file %q: %s", *outFile, err)
	}
	defer out.Close()

	templateData := struct {
		Data      []byte
		Generated time.Time
		RFC3339   string
	}{
		Data:      bytes.Map(sanitize, data),
		Generated: time.Now(),
	}
	if err := template.Execute(out, templateData); err != nil {
		log.Fatalf("failed to execute template: %s", err)
	}

	log.Printf("Successfully wrote %q", *outFile)
}

// sanitize strips characters that can't be `-quoted
func sanitize(r rune) rune {
	switch {
	case r == '`':
		return -1
	case r == '\t', r == '\n':
		return r
	case r >= ' ' && r <= '~':
		return r
	}
	return -1
}
