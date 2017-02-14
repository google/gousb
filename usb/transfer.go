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
#include <libusb.h>

int compact_iso_data(struct libusb_transfer *xfer, unsigned char *status);
int submit(struct libusb_transfer *xfer);
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

//export xfer_callback
func xfer_callback(cptr unsafe.Pointer) {
	ch := *(*chan struct{})(cptr)
	close(ch)
}

type usbTransfer struct {
	xfer *C.struct_libusb_transfer
	buf  []byte
	done chan struct{}
}

type deviceHandle *C.libusb_device_handle

func (t *usbTransfer) attach(dev deviceHandle) {
	t.xfer.dev_handle = dev
}

func (t *usbTransfer) submit() error {
	t.done = make(chan struct{})
	t.xfer.user_data = (unsafe.Pointer)(&t.done)
	if errno := C.submit(t.xfer); errno < 0 {
		return usbError(errno)
	}
	return nil
}

func (t *usbTransfer) wait() (n int, err error) {
	select {
	case <-time.After(10 * time.Second):
		return 0, fmt.Errorf("wait timed out after 10s")
	case <-t.done:
	}
	var status TransferStatus
	switch TransferType(t.xfer._type) {
	case TRANSFER_TYPE_ISOCHRONOUS:
		n = int(C.compact_iso_data(t.xfer, (*C.uchar)(unsafe.Pointer(&status))))
	default:
		n = int(t.xfer.length)
		status = TransferStatus(t.xfer.status)
	}
	if status != LIBUSB_TRANSFER_COMPLETED {
		return 0, status
	}
	return n, err
}

func (t *usbTransfer) free() error {
	C.libusb_free_transfer(t.xfer)
	return nil
}

func newUSBTransfer(ei EndpointInfo, buf []byte, timeout time.Duration) (*usbTransfer, error) {
	var isoPackets int
	tt := ei.TransferType()
	if tt == TRANSFER_TYPE_ISOCHRONOUS {
		isoPackets = len(buf) / int(ei.MaxIsoPacket)
	}

	xfer := C.libusb_alloc_transfer(C.int(isoPackets))
	if xfer == nil {
		return nil, fmt.Errorf("libusb_alloc_transfer(%d) failed", isoPackets)
	}

	xfer.timeout = C.uint(timeout / time.Millisecond)
	xfer.endpoint = C.uchar(ei.Address)
	xfer._type = C.uchar(tt)

	xfer.buffer = (*C.uchar)((unsafe.Pointer)(&buf[0]))
	xfer.length = C.int(len(buf))

	if tt == TRANSFER_TYPE_ISOCHRONOUS {
		xfer.num_iso_packets = C.int(isoPackets)
		C.libusb_set_iso_packet_lengths(xfer, C.uint(ei.MaxIsoPacket))
	}

	return &usbTransfer{
		xfer: xfer,
		buf:  buf,
	}, nil
}
