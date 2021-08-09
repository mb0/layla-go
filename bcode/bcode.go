package bcode

import (
	"fmt"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/ean"
	"github.com/boombuler/barcode/qr"
	"xelf.org/layla"
)

func Barcode(d *layla.Node) (barcode.Barcode, error) {
	switch d.Code.Name {
	case "ean128":
		return code128.Encode(d.Data)
	case "ean8", "ean13":
		return ean.Encode(d.Data)
	}
	if d.Kind != "qrcode" {
		return nil, fmt.Errorf("unknown code name %q", d.Code.Name)
	}
	ec := ErrorCorrection(d.Code.Name)
	return qr.Encode(d.Data, ec, qr.Auto)
}

func ErrorCorrection(name string) qr.ErrorCorrectionLevel {
	switch name {
	case "l":
		return qr.L
	case "m":
		return qr.M
	case "q":
		return qr.Q
	}
	return qr.H
}
