package usb

/*
#include <libusb-1.0/libusb.h>

libusb_transfer_cb_fn callback();
*/
import "C"

import (
	"log"
	"reflect"
	"runtime"
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
		iso_packets = 242
		packet_size = 1760
	)

	xfer := C.libusb_alloc_transfer(C.int(iso_packets))
	if xfer == nil {
		log.Printf("usb: transfer allocation failed?!")
		return nil
	}

	xfer.dev_handle = end.Device.handle
	xfer.endpoint = C.uchar(end.Address)
	xfer._type = C.LIBUSB_TRANSFER_TYPE_ISOCHRONOUS

	xfer.callback = C.callback()
	xfer.timeout = C.uint(5 * time.Second / time.Millisecond)

	buf := make([]byte, iso_packets*packet_size)
	xfer.num_iso_packets = iso_packets
	xfer.buffer = (*C.uchar)((unsafe.Pointer)(&buf[0]))
	xfer.length = C.int(len(buf))
	C.libusb_set_iso_packet_lengths(xfer, C.uint(len(buf)))
	pkts := *(*[]C.struct_libusb_packet_descriptor)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&xfer.iso_packet_desc)),
		Len:  iso_packets,
		Cap:  iso_packets,
	}))

	done := make(chan struct{})
	xfer.user_data = (unsafe.Pointer)(&done)

	t := &Transfer{
		xfer: xfer,
		pkts: pkts,
		done: done,
		buf:  buf,
	}

	runtime.SetFinalizer(t, (*Transfer).Close)

	return t
}

type Transfer struct {
	xfer *C.struct_libusb_transfer
	pkts []C.struct_libusb_packet_descriptor
	done chan struct{}
	buf  []byte
}

func (xfer *Transfer) Submit() error {
	if errno := C.libusb_submit_transfer(xfer.xfer); errno < 0 {
		return usbError(errno)
	}
	return nil
}

func (xfer *Transfer) Wait() (n int, err error) {
	<-xfer.done
	n = int(xfer.xfer.actual_length)
	for _, pkt := range xfer.pkts {
		log.Printf("PACKET[%4d] - %#v", pkt)
	}
	return n, err
}

func (xfer *Transfer) Close() error {
	C.libusb_free_transfer(xfer.xfer)
	return nil
}

func isochronous_xfer(e *endpoint, buf []byte, timeout time.Duration) (int, error) {
	xfer := e.allocTransfer()
	defer xfer.Close()

	if err := xfer.Submit(); err != nil {
		return 0, err
	}

	return xfer.Wait()
}
