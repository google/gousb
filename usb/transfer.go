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

int extract_data(struct libusb_transfer *xfer, void *data, int max, unsigned char *status);
int submit(struct libusb_transfer *xfer);
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

type usbTransfer struct {
	xfer *C.struct_libusb_transfer
	pkts []*C.struct_libusb_packet_descriptor
	done chan struct{}
	buf  []byte
}

func (t *usbTransfer) submit(timeout time.Duration) error {
	t.xfer.timeout = C.uint(timeout / time.Millisecond)
	if errno := C.submit(t.xfer); errno < 0 {
		return usbError(errno)
	}
	return nil
}

func (t *usbTransfer) wait(b []byte) (n int, err error) {
	select {
	case <-time.After(10 * time.Second):
		return 0, fmt.Errorf("wait timed out after 10s")
	case <-t.done:
	}
	var status uint8
	n = int(C.extract_data(t.xfer, unsafe.Pointer(&b[0]), C.int(len(b)), (*C.uchar)(unsafe.Pointer(&status))))
	if status != 0 {
		err = TransferStatus(status)
	}
	return n, err
}

func (t *usbTransfer) free() error {
	C.libusb_free_transfer(t.xfer)
	return nil
}
