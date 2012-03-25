package usb

// #cgo LDFLAGS: -lusb-1.0
// #include <libusb-1.0/libusb.h>
import "C"

type Descriptor struct {
	desc      *C.struct_libusb_device_descriptor

	Type      DescriptorType // The type of this descriptor
	Spec      BCD            // USB Specification Release Number
	Class     Class // The class of this device
	SubClass  uint8 // The sub-class (within the class) of this device
	Protocol  uint8 // The protocol (within the sub-class) of this device
	Vendor    ID    // The 8-bit Vendor identifer
	Product   ID    // The 8-bit Product identifier
	Device    BCD    // The device version
	Configs   int    // Number of configurations
}

func newDescriptor(dev *C.libusb_device) (*Descriptor, error) {
	var desc C.struct_libusb_device_descriptor
	if errno := C.libusb_get_device_descriptor(dev, &desc); errno < 0 {
		return nil, usbError(errno)
	}
	return &Descriptor{
		desc:     &desc,
		Type:     DescriptorType(desc.bDescriptorType),
		Spec:     BCD(desc.bcdUSB),
		Class:    Class(desc.bDeviceClass),
		SubClass: uint8(desc.bDeviceSubClass),
		Protocol: uint8(desc.bDeviceProtocol),
		Vendor:   ID(desc.idVendor),
		Product:  ID(desc.idProduct),
		Device:   BCD(desc.bcdDevice),
		Configs:  int(desc.bNumConfigurations),
	}, nil
}

/*
func (d Descriptor) str(idx int) (string, error) {
	str := [64]byte{}
	n := C.libusb_get_string_descriptor_ascii(
		d.dev.handle,
		C.uint8_t(idx),
		(*C.uchar)(unsafe.Pointer(&str)),
		64,
	)
	if n < 0 {
		return "", usbError(n)
	}
	return string(str[:n]), nil
}
*/

type Class int

const (
	CLASS_PER_INTERFACE Class = C.LIBUSB_CLASS_PER_INTERFACE
	CLASS_AUDIO         Class = C.LIBUSB_CLASS_AUDIO
	CLASS_COMM          Class = C.LIBUSB_CLASS_COMM
	CLASS_HID           Class = C.LIBUSB_CLASS_HID
	CLASS_PRINTER       Class = C.LIBUSB_CLASS_PRINTER
	CLASS_PTP           Class = C.LIBUSB_CLASS_PTP
	CLASS_MASS_STORAGE  Class = C.LIBUSB_CLASS_MASS_STORAGE
	CLASS_HUB           Class = C.LIBUSB_CLASS_HUB
	CLASS_DATA          Class = C.LIBUSB_CLASS_DATA
	CLASS_WIRELESS      Class = C.LIBUSB_CLASS_WIRELESS
	CLASS_APPLICATION   Class = C.LIBUSB_CLASS_APPLICATION
	CLASS_VENDOR_SPEC   Class = C.LIBUSB_CLASS_VENDOR_SPEC
)

var classDescription = map[Class]string{
	CLASS_PER_INTERFACE: "per-interface",
	CLASS_AUDIO:         "audio",
	CLASS_COMM:          "communications",
	CLASS_HID:           "human interface device",
	CLASS_PRINTER:       "printer dclass",
	CLASS_PTP:           "picture transfer protocol",
	CLASS_MASS_STORAGE:  "mass storage",
	CLASS_HUB:           "hub",
	CLASS_DATA:          "data",
	CLASS_WIRELESS:      "wireless",
	CLASS_APPLICATION:   "application",
	CLASS_VENDOR_SPEC:   "vendor-specific",
}

func (c Class) String() string {
	return classDescription[c]
}

type DescriptorType int

const (
	DT_DEVICE    DescriptorType = C.LIBUSB_DT_DEVICE
	DT_CONFIG    DescriptorType = C.LIBUSB_DT_CONFIG
	DT_STRING    DescriptorType = C.LIBUSB_DT_STRING
	DT_INTERFACE DescriptorType = C.LIBUSB_DT_INTERFACE
	DT_ENDPOINT  DescriptorType = C.LIBUSB_DT_ENDPOINT
	DT_HID       DescriptorType = C.LIBUSB_DT_HID
	DT_REPORT    DescriptorType = C.LIBUSB_DT_REPORT
	DT_PHYSICAL  DescriptorType = C.LIBUSB_DT_PHYSICAL
	DT_HUB       DescriptorType = C.LIBUSB_DT_HUB
)

var descriptorTypeDescription = map[DescriptorType]string{
	DT_DEVICE:    "device",
	DT_CONFIG:    "configuration",
	DT_STRING:    "string",
	DT_INTERFACE: "interface",
	DT_ENDPOINT:  "endpoint",
	DT_HID:       "HID",
	DT_REPORT:    "HID report",
	DT_PHYSICAL:  "physical",
	DT_HUB:       "hub",
}

func (dt DescriptorType) String() string {
	return descriptorTypeDescription[dt]
}

type EndpointDirection int

const (
	ENDPOINT_IN  EndpointDirection = C.LIBUSB_ENDPOINT_IN
	ENDPOINT_OUT EndpointDirection = C.LIBUSB_ENDPOINT_OUT
)

var endpointDirectionDescription = map[EndpointDirection]string{
	ENDPOINT_IN:  "device-to-host",
	ENDPOINT_OUT: "host-to-device",
}

func (ed EndpointDirection) String() string {
	return endpointDirectionDescription[ed]
}

type TransferType int

const (
	TRANSFER_TYPE_CONTROL     TransferType = C.LIBUSB_TRANSFER_TYPE_CONTROL
	TRANSFER_TYPE_ISOCHRONOUS TransferType = C.LIBUSB_TRANSFER_TYPE_ISOCHRONOUS
	TRANSFER_TYPE_BULK        TransferType = C.LIBUSB_TRANSFER_TYPE_BULK
	TRANSFER_TYPE_INTERRUPT   TransferType = C.LIBUSB_TRANSFER_TYPE_INTERRUPT
)

var transferTypeDescription = map[TransferType]string{
	TRANSFER_TYPE_CONTROL:     "control",
	TRANSFER_TYPE_ISOCHRONOUS: "isochronous",
	TRANSFER_TYPE_BULK:        "bulk",
	TRANSFER_TYPE_INTERRUPT:   "interrupt",
}

func (tt TransferType) String() string {
	return transferTypeDescription[tt]
}

type IsoSyncType int

const (
	ISO_SYNC_TYPE_NONE     IsoSyncType = C.LIBUSB_ISO_SYNC_TYPE_NONE
	ISO_SYNC_TYPE_ASYNC    IsoSyncType = C.LIBUSB_ISO_SYNC_TYPE_ASYNC
	ISO_SYNC_TYPE_ADAPTIVE IsoSyncType = C.LIBUSB_ISO_SYNC_TYPE_ADAPTIVE
	ISO_SYNC_TYPE_SYNC     IsoSyncType = C.LIBUSB_ISO_SYNC_TYPE_SYNC
)

var isoSyncTypeDescription = map[IsoSyncType]string{
	ISO_SYNC_TYPE_NONE:     "no synchronization",
	ISO_SYNC_TYPE_ASYNC:    "asynchronous",
	ISO_SYNC_TYPE_ADAPTIVE: "adaptive",
	ISO_SYNC_TYPE_SYNC:     "synchronous",
}

func (ist IsoSyncType) String() string {
	return isoSyncTypeDescription[ist]
}

type IsoUsageType int

const (
	ISO_USAGE_TYPE_DATA     IsoUsageType = C.LIBUSB_ISO_USAGE_TYPE_DATA
	ISO_USAGE_TYPE_FEEDBACK IsoUsageType = C.LIBUSB_ISO_USAGE_TYPE_FEEDBACK
	ISO_USAGE_TYPE_IMPLICIT IsoUsageType = C.LIBUSB_ISO_USAGE_TYPE_IMPLICIT
)

var isoUsageTypeDescription = map[IsoUsageType]string{
	ISO_USAGE_TYPE_DATA:     "data",
	ISO_USAGE_TYPE_FEEDBACK: "feedback",
	ISO_USAGE_TYPE_IMPLICIT: "implicit data",
}

func (iut IsoUsageType) String() string {
	return isoUsageTypeDescription[iut]
}
