package usbid

import (
	"log"
	"net/http"
	"strings"

	"github.com/kylelemons/gousb/usb"
)

const (
	// LinuxUsbDotOrg is one source of files in the format used by this package.
	LinuxUsbDotOrg = "http://www.linux-usb.org/usb.ids"
)

var (
	// Vendors stores the vendor and product ID mappings.
	Vendors map[usb.ID]*Vendor

	// Classes stores the class, subclass and protocol mappings.
	Classes map[uint8]*Class
)

// LoadFromURL replaces the built-in vendor and class mappings with ones loaded
// from the given URL.
//
// This should usually only be necessary if the mappings in the library are
// stale.  The contents of this file as of February 2012 are embedded in the
// library itself.
func LoadFromURL(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ids, cls, err := ParseIDs(resp.Body)
	if err != nil {
		return err
	}

	Vendors = ids
	Classes = cls
	return nil
}

func init() {
	ids, cls, err := ParseIDs(strings.NewReader(usbIdListData))
	if err != nil {
		log.Printf("usbid: failed to parse: %s", err)
		return
	}

	Vendors = ids
	Classes = cls
}
