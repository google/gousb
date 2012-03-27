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
	Product map[usb.ID]*Product
}

func (v Vendor) String() string {
	return v.Name
}

type Product struct {
	Name      string
	Interface map[usb.ID]string
}

func (p Product) String() string {
	return p.Name
}

type Class struct {
	Name     string
	SubClass map[uint8]*SubClass
}

func (c Class) String() string {
	return c.Name
}

type SubClass struct {
	Name     string
	Protocol map[uint8]string
}

func (s SubClass) String() string {
	return s.Name
}

func ParseIDs(r io.Reader) (map[usb.ID]*Vendor, map[uint8]*Class, error) {
	vendors := make(map[usb.ID]*Vendor, 2800)
	classes := make(map[uint8]*Class) // TODO(kevlar): count

	split := func(s string) (kind string, level int, id uint64, name string, err error) {
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
		id = i

		return
	}

	// Hold the interim values
	var vendor *Vendor
	var device *Product

	parseVendor := func(level int, raw uint64, name string) error {
		id := usb.ID(raw)

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

			device = &Product{
				Name: name,
			}
			if vendor.Product == nil {
				vendor.Product = make(map[usb.ID]*Product)
			}
			vendor.Product[id] = device

		case 2:
			if device == nil {
				return fmt.Errorf("interface line without device line")
			}

			if device.Interface == nil {
				device.Interface = make(map[usb.ID]string)
			}
			device.Interface[id] = name

		default:
			return fmt.Errorf("too many levels of nesting for vendor block")
		}

		return nil
	}

	// Hold the interim values
	var class *Class
	var subclass *SubClass

	parseClass := func(level int, raw uint64, name string) error {
		id := uint8(raw)

		switch level {
		case 0:
			class = &Class{
				Name: name,
			}
			classes[id] = class

		case 1:
			if class == nil {
				return fmt.Errorf("subclass line without class line")
			}

			subclass = &SubClass{
				Name: name,
			}
			if class.SubClass == nil {
				class.SubClass = make(map[uint8]*SubClass)
			}
			class.SubClass[id] = subclass

		case 2:
			if subclass == nil {
				return fmt.Errorf("protocol line without subclass line")
			}

			if subclass.Protocol == nil {
				subclass.Protocol = make(map[uint8]string)
			}
			subclass.Protocol[id] = name

		default:
			return fmt.Errorf("too many levels of nesting for class")
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
			return nil, nil, err
		case isPrefix:
			return nil, nil, fmt.Errorf("line %d: line too long", lineno)
		}
		line := string(b)

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		k, level, id, name, err := split(line)
		if err != nil {
			return nil, nil, fmt.Errorf("line %d: %s", lineno, err)
		}
		if k != "" {
			kind = k
		}

		switch kind {
		case "":
			err = parseVendor(level, id, name)
		case "C":
			err = parseClass(level, id, name)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("line %d: %s", lineno, err)
		}
	}

	return vendors, classes, nil
}
