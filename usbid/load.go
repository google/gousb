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

func LoadFromURL(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ids, err := ParseIDs(resp.Body)
	if err != nil {
		return err
	}

	Vendors = ids
	return nil
}

func init() {
	ids, err := ParseIDs(strings.NewReader(usbIdListData))
	if err != nil {
		log.Printf("usbid: failed to parse: %s", err)
		return
	}

	Vendors = ids
}
