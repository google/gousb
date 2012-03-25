package usbid

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/kylelemons/gousb/usb"
)

type Vendor struct {
	Name    string
	Devices map[usb.ID]*Device
}

func (v Vendor) String() string {
	return v.Name
}

// TODO(kevlar) s/device/
type Device struct {
	Name       string
	Interfaces map[usb.ID]string
}

func (d Device) String() string {
	return d.Name
}

/*
type Class struct {
	Name string
	SubClass map[usb.ID]*SubClass
}

type SubClass struct {
	Name string
	Protocol map[usb.ID]string
}
*/

func ParseIDs(r io.Reader) (map[usb.ID]*Vendor, error) {
	vendors := make(map[usb.ID]*Vendor, 2800)

	split := func(s string) (kind string, level int, id usb.ID, name string, err error) {
		pieces := strings.SplitN(s, "  ", 2)
		if len(pieces) != 2 {
			err = fmt.Errorf("malformatted line %q", s)
			return
		}

		// Save the name
		name = pieces[1]

		// Parse out the level
		for len(pieces[0]) > 0 && pieces[0][0] == '\t' {
			level, pieces[0] = level+1, pieces[0][1:]
		}

		// Parse the first piece to see if it has a kind
		first := strings.SplitN(pieces[0], " ", 2)
		if len(first) == 2 {
			kind, pieces[0] = first[0], first[1]
		}

		// Parse the ID
		i, err := strconv.ParseUint(pieces[0], 16, 16)
		if err != nil {
			err = fmt.Errorf("malformatted id %q: %s", pieces[0], err)
			return
		}
		id = usb.ID(i)
		
		return
	}

	// Hold the interim values
	var vendor *Vendor
	var device *Device

	parseVendor := func(level int, id usb.ID, name string) error {
		switch level {
		case 0:
			vendor = &Vendor{
				Name: name,
			}
			vendors[id] = vendor

		case 1:
			if vendor == nil {
				return fmt.Errorf("product line without vendor line")
			}

			device = &Device{
				Name: name,
			}
			if vendor.Devices == nil {
				vendor.Devices = make(map[usb.ID]*Device)
			}
			vendor.Devices[id] = device

		case 2:
			if device == nil {
				return fmt.Errorf("interface line without device line")
			}

			if device.Interfaces == nil {
				device.Interfaces = make(map[usb.ID]string)
			}
			device.Interfaces[id] = name

		default:
			return fmt.Errorf("too many levels of nesting for vendor block")
		}

		return nil
	}

	// TODO(kevlar): Parse class information, etc
	//var class *Class
	//var subclass *SubClass

	var kind string

	lines := bufio.NewReaderSize(r, 512)
parseLines:
	for lineno := 0; ; lineno++ {
		b, isPrefix, err := lines.ReadLine()
		switch {
		case err == io.EOF:
			break parseLines
		case err != nil:
			return nil, err
		case isPrefix:
			return nil, fmt.Errorf("line %d: line too long", lineno)
		}
		line := string(b)

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		k, level, id, name, err := split(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %s", lineno, err)
		}
		if k != "" {
			kind = k
		}

		switch kind {
		case "": err = parseVendor(level, id, name)
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %s", lineno, err)
		}
	}

	return vendors, nil
}
