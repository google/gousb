package usb

// #include <libusb-1.0/libusb.h>
import "C"

type usbError C.int

func (e usbError) Error() string {
	return "libusb: " + usbErrorString[e]
}

const (
	SUCCESS             usbError = C.LIBUSB_SUCCESS
	ERROR_IO            usbError = C.LIBUSB_ERROR_IO
	ERROR_INVALID_PARAM usbError = C.LIBUSB_ERROR_INVALID_PARAM
	ERROR_ACCESS        usbError = C.LIBUSB_ERROR_ACCESS
	ERROR_NO_DEVICE     usbError = C.LIBUSB_ERROR_NO_DEVICE
	ERROR_NOT_FOUND     usbError = C.LIBUSB_ERROR_NOT_FOUND
	ERROR_BUSY          usbError = C.LIBUSB_ERROR_BUSY
	ERROR_TIMEOUT       usbError = C.LIBUSB_ERROR_TIMEOUT
	ERROR_OVERFLOW      usbError = C.LIBUSB_ERROR_OVERFLOW
	ERROR_PIPE          usbError = C.LIBUSB_ERROR_PIPE
	ERROR_INTERRUPTED   usbError = C.LIBUSB_ERROR_INTERRUPTED
	ERROR_NO_MEM        usbError = C.LIBUSB_ERROR_NO_MEM
	ERROR_NOT_SUPPORTED usbError = C.LIBUSB_ERROR_NOT_SUPPORTED
	ERROR_OTHER         usbError = C.LIBUSB_ERROR_OTHER
)

var usbErrorString = map[usbError]string{
	C.LIBUSB_SUCCESS:             "success",
	C.LIBUSB_ERROR_IO:            "i/o error",
	C.LIBUSB_ERROR_INVALID_PARAM: "invalid param",
	C.LIBUSB_ERROR_ACCESS:        "bad access",
	C.LIBUSB_ERROR_NO_DEVICE:     "no device",
	C.LIBUSB_ERROR_NOT_FOUND:     "not found",
	C.LIBUSB_ERROR_BUSY:          "device or resource busy",
	C.LIBUSB_ERROR_TIMEOUT:       "timeout",
	C.LIBUSB_ERROR_OVERFLOW:      "overflow",
	C.LIBUSB_ERROR_PIPE:          "pipe error",
	C.LIBUSB_ERROR_INTERRUPTED:   "interrupted",
	C.LIBUSB_ERROR_NO_MEM:        "out of memory",
	C.LIBUSB_ERROR_NOT_SUPPORTED: "not supported",
	C.LIBUSB_ERROR_OTHER:         "unknown error",
}
