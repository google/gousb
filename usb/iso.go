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
	"unsafe"
)

//export iso_callback
func iso_callback(cptr unsafe.Pointer) {
	ch := *(*chan struct{})(cptr)
	close(ch)
}

func (end *endpoint) allocTransfer() *usbTransfer {
	const (
		iso_packets = 8 // 128 // 242
	)

	xfer := C.libusb_alloc_transfer(C.int(iso_packets))
	if xfer == nil {
		log.Printf("usb: transfer allocation failed?!")
		return nil
	}

	buf := make([]byte, iso_packets*end.EndpointInfo.MaxIsoPacket)
	done := make(chan struct{}, 1)

	xfer.dev_handle = end.Device.handle
	xfer.endpoint = C.uchar(end.Address)
	xfer._type = C.LIBUSB_TRANSFER_TYPE_ISOCHRONOUS

	xfer.buffer = (*C.uchar)((unsafe.Pointer)(&buf[0]))
	xfer.length = C.int(len(buf))
	xfer.num_iso_packets = iso_packets

	C.libusb_set_iso_packet_lengths(xfer, C.uint(end.EndpointInfo.MaxIsoPacket))
	/*
		pkts := *(*[]C.struct_libusb_packet_descriptor)(unsafe.Pointer(&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(&xfer.iso_packet_desc)),
			Len:  iso_packets,
			Cap:  iso_packets,
		}))
	*/

	t := &usbTransfer{
		xfer: xfer,
		done: done,
		buf:  buf,
	}
	xfer.user_data = (unsafe.Pointer)(&t.done)

	return t
}

func isochronous_xfer(e *endpoint, buf []byte, timeout time.Duration) (int, error) {
	t := e.allocTransfer()
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
