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

package usb

import (
	"testing"
	"time"
	"unsafe"
)

func TestNewTransfer(t *testing.T) {
	for _, tc := range []struct {
		desc        string
		dir         EndpointDirection
		tt          TransferType
		maxPkt      uint16
		maxIso      uint32
		buf         int
		timeout     time.Duration
		wantIso     int
		wantLength  int
		wantTimeout int
	}{
		{
			desc:        "bulk in transfer, 512B packets",
			dir:         ENDPOINT_DIR_IN,
			tt:          TRANSFER_TYPE_BULK,
			maxPkt:      512,
			buf:         1024,
			timeout:     time.Second,
			wantIso:     0,
			wantLength:  1024,
			wantTimeout: 1000,
		},
		{
			desc:        "iso out transfer, 3 * 1024B packets",
			dir:         ENDPOINT_DIR_OUT,
			tt:          TRANSFER_TYPE_ISOCHRONOUS,
			maxPkt:      2<<11 + 1024,
			maxIso:      3 * 1024,
			buf:         10000,
			timeout:     500 * time.Millisecond,
			wantIso:     4,
			wantLength:  10000,
			wantTimeout: 500,
		},
	} {
		xfer, err := newUSBTransfer(nil, EndpointInfo{
			Address:       uint8(tc.dir) | 0x02,
			Attributes:    uint8(tc.tt),
			MaxPacketSize: tc.maxPkt,
			MaxIsoPacket:  tc.maxIso,
			PollInterval:  1,
		}, make([]byte, tc.buf), tc.timeout)

		if err != nil {
			t.Fatalf("newUSBTransfer(): %v", err)
		}
		if got, want := len(xfer.buf), tc.buf; got != want {
			t.Errorf("xfer.buf: got %d bytes, want %d", got, want)
		}
		if got, want := int(xfer.xfer.length), tc.buf; got != want {
			t.Errorf("xfer.length: got %d, want %d", got, want)
		}
		if got, want := int(xfer.xfer.num_iso_packets), tc.wantIso; got != want {
			t.Errorf("xfer.num_iso_packets: got %d, want %d", got, want)
		}
		if got, want := int(xfer.xfer.timeout), tc.wantTimeout; got != want {
			t.Errorf("xfer.timeout: got %d ms, want %d", got, want)
		}
		if got, want := TransferType(xfer.xfer._type), tc.tt; got != want {
			t.Errorf("xfer._type: got %s, want %s", got, want)
		}
	}
}

func TestTransferProtocol(t *testing.T) {
	origSubmit, origCancel := cSubmit, cCancel
	defer func() { cSubmit, cCancel = origSubmit, origCancel }()

	f := newFakeLibusb()
	cSubmit = f.submit
	cCancel = f.cancel

	xfers := make([]*usbTransfer, 2)
	var err error
	for i := 0; i < 2; i++ {
		xfers[i], err = newUSBTransfer(nil, EndpointInfo{
			Address:       0x86,
			Attributes:    uint8(TRANSFER_TYPE_BULK),
			MaxPacketSize: 512,
			PollInterval:  1,
		}, make([]byte, 10240), time.Second)
		if err != nil {
			t.Fatalf("newUSBTransfer: %v", err)
		}
	}

	go func() {
		f.waitForSubmit(xfers[0])
		f.runCallback(xfers[0], func(t *usbTransfer) {
			t.xfer.actual_length = 5
			t.xfer.status = uint32(SUCCESS)
			copy(t.buf, []byte{1, 2, 3, 4, 5})
		})
	}()
	go func() {
		f.waitForSubmit(xfers[1])
		f.runCallback(xfers[1], func(t *usbTransfer) {
			t.xfer.actual_length = 99
			t.xfer.status = uint32(SUCCESS)
			copy(t.buf, []byte{12, 12, 12, 12, 12})
		})
	}()

	xfers[1].submit()
	xfers[0].submit()
	got, err := xfers[0].wait()
	if err != nil {
		t.Errorf("xfer#0.wait returned error %v, want nil", err)
	}
	if want := 5; got != want {
		t.Errorf("xfer#0.wait returned %d bytes, want %d", got, want)
	}
	got, err = xfers[1].wait()
	if err != nil {
		t.Errorf("xfer#0.wait returned error %v, want nil", err)
	}
	if want := 99; got != want {
		t.Errorf("xfer#0.wait returned %d bytes, want %d", got, want)
	}

	go func() {
		f.waitForSubmit(xfers[1])
		f.runCallback(xfers[1], func(t *usbTransfer) {
			t.xfer.actual_length = 123
			t.xfer.status = uint32(LIBUSB_TRANSFER_CANCELLED)
		})
	}()
	xfers[1].submit()
	xfers[1].cancel()
	got, err = xfers[1].wait()
	if err == nil {
		t.Error("xfer#1(resubmitted).wait returned error nil, want non-nil")
	}
	if want := 123; got != want {
		t.Errorf("xfer#1(resubmitted).wait returned %d bytes, want %d", got, want)
	}

	for i := 0; i < 2; i++ {
		xfers[i].cancel()
		xfers[i].wait()
		xfers[i].free()
	}
}

func TestIsoPackets(t *testing.T) {
	origSubmit, origCancel := cSubmit, cCancel
	defer func() { cSubmit, cCancel = origSubmit, origCancel }()

	f := newFakeLibusb()
	cSubmit = f.submit
	cCancel = f.cancel

	xfer, err := newUSBTransfer(nil, EndpointInfo{
		Address:       0x82,
		Attributes:    uint8(TRANSFER_TYPE_ISOCHRONOUS),
		MaxPacketSize: 3<<11 + 1024,
		MaxIsoPacket:  3 * 1024,
		PollInterval:  1,
	}, make([]byte, 15000), time.Second)
	if err != nil {
		t.Fatalf("newUSBTransfer: %v", err)
	}

	// 15000 / (3*1024) = 4.something, rounded up to 5
	if got, want := int(xfer.xfer.num_iso_packets), 5; got != want {
		t.Fatalf("newUSBTransfer: got %d iso packets, want %d", got, want)
	}

	go func() {
		f.waitForSubmit(xfer)
		f.runCallback(xfer, func(x *usbTransfer) {
			x.xfer.actual_length = 1234 // meaningless for iso transfers
			x.xfer.status = uint32(LIBUSB_TRANSFER_TIMED_OUT)
			for i := 0; i < int(xfer.xfer.num_iso_packets); i++ {
				// this is a horrible calculation.
				// libusb_transfer uses a flexible array for the iso packet
				// descriptors at the end of the transfer struct.
				// The only way to get access to the elements of that array
				// is to use pointer arithmetic.
				// Calculate the offset of the first descriptor in the struct,
				// then move by sizeof(iso descriptor) for num_iso_packets.
				desc := (*libusbIso)(unsafe.Pointer(uintptr(unsafe.Pointer(x.xfer)) + libusbIsoOffset + uintptr(i*libusbIsoSize)))
				// max iso packet = 3 * 1024
				if desc.length != 3*1024 {
					t.Errorf("iso pkt length: got %d, want %d", desc.length, 3*1024)
				}
				desc.actual_length = 100
				// packets 0..2 are successful, packet 3 is timed out
				if i != 4 {
					desc.status = uint32(LIBUSB_TRANSFER_COMPLETED)
				} else {
					desc.status = uint32(LIBUSB_TRANSFER_TIMED_OUT)
				}
			}
		})
	}()

	xfer.submit()
	got, err := xfer.wait()
	if err == nil {
		t.Error("Iso transfer: got nil error, want non-nil")
	}
	if want := 4 * 100; got != want {
		t.Errorf("Iso transfer: got %d bytes, want %d", got, want)
	}
}
