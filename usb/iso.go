package usb

/*
#include <libusb-1.0/libusb.h>

int submit(struct libusb_transfer *xfer);
*/
import "C"

import (
	"log"
	"fmt"
	"reflect"
	"time"
	"unsafe"
)

//export iso_callback
func iso_callback(cptr unsafe.Pointer) {
	ch := *(*chan struct{})(cptr)
	close(ch)
}

func (end *endpoint) allocTransfer() *Transfer {
	// Use libusb_get_max_iso_packet_size ?
	const (
		iso_packets = 8 // 128 // 242
		packet_size = 2*960 // 1760
	)

	xfer := C.libusb_alloc_transfer(C.int(iso_packets))
	if xfer == nil {
		log.Printf("usb: transfer allocation failed?!")
		return nil
	}

	buf := make([]byte, iso_packets*packet_size)
	done := make(chan struct{}, 1)

	xfer.dev_handle = end.Device.handle
	xfer.endpoint = C.uchar(end.Address)
	xfer._type = C.LIBUSB_TRANSFER_TYPE_ISOCHRONOUS

	xfer.buffer = (*C.uchar)((unsafe.Pointer)(&buf[0]))
	xfer.length = C.int(len(buf))
	xfer.num_iso_packets = iso_packets

	xfer.user_data = (unsafe.Pointer)(&done)

	C.libusb_set_iso_packet_lengths(xfer, packet_size)
	pkts := *(*[]C.struct_libusb_packet_descriptor)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&xfer.iso_packet_desc)),
		Len:  iso_packets,
		Cap:  iso_packets,
	}))

	t := &Transfer{
		xfer: xfer,
		pkts: pkts,
		done: done,
		buf:  buf,
	}

	return t
}

type Transfer struct {
	xfer *C.struct_libusb_transfer
	pkts []C.struct_libusb_packet_descriptor
	done chan struct{}
	buf  []byte
}

func (t *Transfer) Submit(timeout time.Duration) error {
	log.Printf("iso: submitting %#v", t.xfer)
	t.xfer.timeout = C.uint(timeout / time.Millisecond)
	if errno := C.submit(t.xfer); errno < 0 {
		return usbError(errno)
	}
	return nil
}

func (t *Transfer) Wait(b []byte) (n int, err error) {
	select {
	case <-time.After(10*time.Second):
		return 0, fmt.Errorf("wait timed out after 10s")
	case <-t.done:
	}
	n = int(t.xfer.actual_length)
	copy(b, ((*[1<<16]byte)(unsafe.Pointer(t.xfer.buffer)))[:n])
	/*
	for i, pkt := range t.pkts {
		log.Printf("PACKET[%4d] - %#v", i, pkt)
	}*/
	return n, err
}

func (t *Transfer) Close() error {
	C.libusb_free_transfer(t.xfer)
	return nil
}

func isochronous_xfer(e *endpoint, buf []byte, timeout time.Duration) (int, error) {
	t := e.allocTransfer()
	defer t.Close()

	if err := t.Submit(timeout); err != nil {
		log.Printf("iso: xfer failed to submit: %s", err)
		return 0, err
	}

	n, err := t.Wait(buf)
	if err != nil {
		log.Printf("iso: xfer failed: %s", err)
		return 0, err
	}
	return n, err
}
