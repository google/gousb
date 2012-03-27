package usbid

import (
	"log"
	"net/http"
	"strings"

	"github.com/kylelemons/gousb/usb"
)

const (
	LinuxUsbDotOrg = "http://www.linux-usb.org/usb.ids"
)

var Vendors map[usb.ID]*Vendor
var Classes map[uint8]*Class

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
