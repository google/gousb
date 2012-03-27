package usbid

import (
	"fmt"

	"github.com/kylelemons/gousb/usb"
)

// Describe one of the following:
//   - *usb.Descriptor       Product (Vendor)
func Describe(val interface{}) string {
	switch val := val.(type) {
	case *usb.Descriptor:
		if v, ok := Vendors[val.Vendor]; ok {
			if d, ok := v.Product[val.Product]; ok {
				return fmt.Sprintf("%s (%s)", d, v)
			}
			return fmt.Sprintf("Unknown (%s)", v)
		}
		return fmt.Sprintf("Unknown %s:%s", val.Vendor, val.Product)
	}
	return fmt.Sprintf("Unknown (%T)", val)
}

// Classify one of the following:
//   - *usb.Descriptor       Class SubClass Protocol
func Classify(val interface{}) string {
	var class, sub, proto uint8
	switch val := val.(type) {
	case *usb.Descriptor:
		class, sub, proto = val.Class, val.SubClass, val.Protocol
	case *usb.Interface:
		class, sub, proto = val.IfClass, val.IfSubClass, val.IfProtocol
	default:
		return fmt.Sprintf("Unknown (%T)", val)
	}

	if c, ok := Classes[class]; ok {
		if s, ok := c.SubClass[sub]; ok {
			if p, ok := s.Protocol[proto]; ok {
				return fmt.Sprintf("%s (%s) %s", c, s, p)
			}
			return fmt.Sprintf("%s (%s)", c, s)
		}
		return fmt.Sprintf("%s", c)
	}
	return fmt.Sprintf("Unknown %s.%s.%s", class, sub, proto)
}
