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

package usb

/*
#include <libusb-1.0/libusb.h>
*/
import "C"

import (
	"log"
	"time"
)

func isochronous_xfer(e *endpoint, buf []byte, timeout time.Duration) (int, error) {
	t, err := e.newUSBTransfer(TRANSFER_TYPE_ISOCHRONOUS, buf)
	if err != nil {
		return 0, err
	}
	defer t.free()

	if err := t.submit(timeout); err != nil {
		log.Printf("iso: xfer failed to submit: %s", err)
		return 0, err
	}

	n, err := t.wait(buf)
	if err != nil {
		log.Printf("iso: xfer failed: %s", err)
		return 0, err
	}
	return n, err
}
