// Package html implements a layla renderer for html previews.
// Both barcodes and qrcodes are generated as images and embedded as data urls.
package html

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"strings"

	"github.com/boombuler/barcode"
	"xelf.org/layla"
	"xelf.org/layla/bcode"
	"xelf.org/layla/font"
	"xelf.org/layla/mark"
	"xelf.org/xelf/bfr"
)

// Render renders the node n as HTML to b or returns an error.
func Render(b bfr.Writer, man *font.Manager, n *layla.Node) error {
	draw, err := layla.LayoutAndPage(man, n)
	if err != nil {
		return err
	}
	b.WriteString(`<style>
@font-face {
	font-family: 'GoReg.ttf';
	src: url('font/Go-Regular.ttf') format('truetype');
}
@font-face {
	font-family: 'GoBold.ttf';
	src: url('font/Go-Bold.ttf') format('truetype');
}
.layla {
	position: relative;
	background-color: white;
	margin: 10mm;
	text-rendering: optimizeSpeed;
	font-kerning: normal;
}
.layla * {
	position: absolute;
	box-sizing: border-box;
}</style>
`)
	for i, d := range draw {
		if i == 0 || d.Kind == "page" {
			if i > 0 {
				b.WriteString("</div>\n")
			}
			fmt.Fprintf(b, `<div class="layla" style="width:%gmm;height:%gmm">`+"\n", n.W/8, n.H/8)
			if d.Kind == "page" {
				continue
			}
		}
		b.WriteString(`<div style="`)
		switch d.Kind {
		case "ellipse":
			writeBox(b, d.Box, d.Border.W)
			fmt.Fprintf(b, "border:%gmm solid black;", d.Border.W/8)
			b.WriteString(`border-radius: 50%">`)
		case "line":
			if d.W == 0 {
				writeBox(b, d.Box, d.Border.W)
				fmt.Fprintf(b, "border-left:%gmm solid black;", d.Border.W/8)
			} else if d.H == 0 {
				writeBox(b, d.Box, d.Border.W)
				fmt.Fprintf(b, "border-top:%gmm solid black;", d.Border.W/8)
			} else {
				hyp := font.Dot(math.Sqrt(float64(d.W*d.W + d.H*d.H)))
				deg := math.Asin(float64(d.H/hyp)) * 180 / math.Pi
				pos := layla.Pos{d.X + d.Border.W*.25, d.Y - d.Border.W*.5}
				if deg < 0 {
					pos = layla.Pos{d.X - d.Border.W*.25, d.Y}
				}
				writeBox(b, layla.Box{pos, layla.Dim{hyp.Round(), 0}}, 0)
				fmt.Fprintf(b, "border-top:%gmm solid black;", d.Border.W/8)
				fmt.Fprintf(b, "transform:rotate(%gdeg);", math.Round(deg*10)/10)
				b.WriteString(`transform-origin:top left;`)
			}
			b.WriteString(`">`)
		case "rect":
			writeBox(b, d.Box, d.Border.W)
			fmt.Fprintf(b, "border:%gmm solid black;", d.Border.W/8)
			b.WriteString(`">`)
		case "text":
			y, fsize := d.Y, d.Font.Size
			if man.Compat { // tspl render compatibility mode
				// for some reason these parameters fit tspl label printer text rendering
				y -= font.Dot(fsize * .55)
				fsize *= .96
			}
			fmt.Fprintf(b, "left:%gmm;", (d.X-1)/8)
			fmt.Fprintf(b, "top:%gmm;", y/8)
			fmt.Fprintf(b, "width:%gmm;", d.W/8)
			fmt.Fprintf(b, "height:%gmm;", d.H/8)
			fmt.Fprintf(b, "font-family:'%s';", d.Font.Name)
			fmt.Fprintf(b, "font-size:%gpt;", fsize)
			fmt.Fprintf(b, "line-height:%gmm;", d.Font.Line/8)
			if d.Font.Style&mark.Bold != 0 {
				fmt.Fprintf(b, "font-weight:bold;")
			}
			if d.Border.W > 0 {
				fmt.Fprintf(b, "border:%gmm solid black;", d.Border.W/8)
			}
			switch d.Align {
			case 1:
				fmt.Fprintf(b, "text-align:right;")
			case 2:
				fmt.Fprintf(b, "text-align:center;")
			}
			b.WriteString(`">`)
			b.WriteString(strings.ReplaceAll(d.Data, "\n", "<br>\n"))
		case "barcode", "qrcode":
			writeBox(b, d.Box, 0)
			b.WriteString(`">`)
			err = writeBarcode(b, d)
			if err != nil {
				return err
			}
		}
		b.WriteString("</div>\n")
	}
	b.WriteString(`</div>`)
	return nil
}
func writeBox(b bfr.Writer, d layla.Box, border layla.Dot) {
	fmt.Fprintf(b, "left:%gmm;", (d.X-border*.5)/8)
	fmt.Fprintf(b, "top:%gmm;", (d.Y-border*.5)/8)
	fmt.Fprintf(b, "width:%gmm;", (d.W+border)/8)
	fmt.Fprintf(b, "height:%gmm;", (d.H+border)/8)
}

func writeBarcode(b bfr.Writer, d *layla.Node) error {
	img, err := bcode.Barcode(d)
	if err != nil {
		return err
	}
	img, err = barcode.Scale(img, int(d.W), int(d.H))
	if err != nil {
		log.Printf("scale barcode %g %g", d.W, d.H)
		return err
	}
	fmt.Fprintf(b, `<img style="width:%gmm; height:%gmm" src="`, d.W/8, d.H/8)
	err = writeDataURL(b, img)
	if err != nil {
		return err
	}
	fmt.Fprintf(b, `" alt="%s">`, d.Kind)
	return nil
}

// writeDataURL writes the given img as monochrome gif data url to b
func writeDataURL(b bfr.Writer, img image.Image) error {
	b.WriteString("data:image/png;base64,")
	enc := base64.NewEncoder(base64.RawStdEncoding, b)
	err := png.Encode(enc, img)
	if err != nil {
		return err
	}
	return nil
}
