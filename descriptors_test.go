// Copyright 2017 the gousb Authors.  All rights reserved.
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
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func bytesFromHexFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Open(): %v", err)
	}
	defer f.Close()
	hexData, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("ReadAll(): %v", err)
	}
	hexData = bytes.TrimSpace(hexData)
	if len(hexData)%2 == 1 {
		return nil, fmt.Errorf("hex data file has %d characters, expected an even number", len(hexData))
	}
	data := make([]byte, len(hexData)/2)
	_, err = hex.Decode(data, hexData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex data: %v", err)
	}
	return data, nil
}

func TestDeviceDescriptor(t *testing.T) {
	path := "testdata/mouse_device_desc.hex"
	descData, err := bytesFromHexFile(path)
	if err != nil {
		t.Fatalf("loading data from %q failed: %v", path, err)
	}

	got, err := deviceDescFromBytes(descData)
	if err != nil {
		t.Fatalf("failed to decode device descriptor from %q: %v", path, err)
	}

	// based on device descriptor as presented in mouse_lsusb.txt
	want := &DeviceDesc{
		Spec:                 Version(2, 0),
		Class:                ClassPerInterface,
		SubClass:             ClassPerInterface,
		Protocol:             Protocol(0),
		MaxControlPacketSize: 32,
		Vendor:               ID(0x046d),
		Product:              ID(0xc526),
		Device:               Version(5, 0),
		Configs:              make(map[int]ConfigDesc),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got descriptor: %+v\nwant: %+v", got, want)
	}
}
