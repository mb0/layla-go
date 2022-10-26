// Package tspl implements a renderer for TSPL format used by TSC thermal label printer.
package tspl

import (
	"fmt"
	"strings"

	"xelf.org/layla"
	"xelf.org/layla/font"
	"xelf.org/layla/mark"
	"xelf.org/xelf/bfr"
)

// Render renders the node n as TSPL to b or returns an error.
func Render(b bfr.Writer, man *font.Manager, n *layla.Node, extra ...string) error {
	lay := &layla.Layouter{Manager: man, Spacer: 'i', Styler: layla.FakeBoldStyler}
	draw, err := lay.LayoutAndPage(n)
	if err != nil {
		return err
	}
	w, h := n.W, n.H
	if n.Rot == 90 || n.Rot == -90 || n.Rot == 270 {
		w, h = h, w
	}
	fmt.Fprintf(b, "SIZE %g mm, %g mm\n", w/8, h/8)
	fmt.Fprintf(b, "GAP %g mm, 0 mm\n", n.Gap/8)
	b.WriteString("DIRECTION 1,0\nCODEPAGE UTF-8\n")
	for _, line := range extra {
		b.WriteString(line)
		if len(line) > 0 && line[len(line)-1] != '\n' {
			b.WriteByte('\n')
		}
	}
	b.WriteString("CLS\n")
	for _, d := range draw {
		err = renderNode(lay, b, d, n.Rot, n.W, n.H)
		if err != nil {
			return err
		}
	}
	return nil
}

func renderNode(lay *layla.Layouter, b bfr.Writer, d *layla.Node, rot int, rw, rh layla.Dot) error {
	switch rot {
	case 90:
		switch d.Kind {
		case "rect", "line", "ellipse":
			d.X, d.Y = rh-d.Y-d.H, d.X
			d.W, d.H = d.H, d.W
		case "text", "barcode", "qrcode":
			d.X, d.Y = rh-d.Y, d.X
		}
	case -90, 270:
		rot = 270
		switch d.Kind {
		case "rect", "line", "ellipse":
			d.X, d.Y = d.Y, rw-d.X-d.H
			d.W, d.H = d.H, d.W
		case "text", "barcode", "qrcode":
			d.X, d.Y = d.Y-d.H, rw-d.X
		}
	}
	dpi := lay.DPI()
	switch d.Kind {
	case "ellipse":
		w := d.Border.W.At(dpi)
		fmt.Fprintf(b, "ELLIPSE %d,%d,%d,%d,%d\n",
			d.X.At(dpi)-w, d.Y.At(dpi)-w, d.W.At(dpi), d.H.At(dpi), w)
	case "rect":
		w := d.Border.W.At(dpi)
		fmt.Fprintf(b, "BOX %d,%d,%d,%d,%d\n",
			d.X.At(dpi)-w, d.Y.At(dpi)-w, (d.X + d.W).At(dpi), (d.Y + d.H).At(dpi), w)
	case "line":
		fmt.Fprintf(b, "DIAGONAL %d,%d,%d,%d,%d\n",
			d.X.At(dpi), d.Y.At(dpi), (d.X + d.W).At(dpi), (d.Y + d.H).At(dpi),
			d.Border.W.At(dpi))
	case "text":
		fnt := "0"
		if d.Font.Name != "" {
			fnt = strings.ToUpper(d.Font.Name)
		}
		fsize := fontSize(d)
		data := strings.Replace(fmt.Sprintf("%q", d.Data), "\\n", "\\[L]", -1)
		space := (d.Font.Line - lay.PtToDot(d.Font.Height).Ceil()).Floor()
		x, w := d.X.At(dpi), d.W.At(dpi)
		// TODO fix overflow due to discrepancy between font measuring and printing
		// the reason might be that the tsc printer does not apply kerning?
		switch d.Align {
		case 3:
			x -= 10
			w += 10
		case 2:
			x -= 5
			w += 5
		default:
			w += 10
		}
		fmt.Fprintf(b, "BLOCK %d,%d,%d,%d,\"%s\",%d,%d,%d,%d,%d,%s\n",
			x, d.Y.At(dpi), w, d.H.At(dpi), fnt, rot,
			fsize, fsize, space.At(dpi), d.Align, data)
		if d.Font != nil && d.Font.Style&mark.Bold != 0 {
			fmt.Fprintf(b, "BLOCK %d,%d,%d,%d,\"%s\",%d,%d,%d,%d,%d,%s\n",
				x+1, d.Y.At(dpi), w+1, d.H.At(dpi), fnt, rot,
				fsize, fsize, space.At(dpi), d.Align, data)
		}
	case "barcode":
		h := d.H.At(dpi)
		if d.Code.Human != 0 {
			h -= font.Dot(20).At(dpi)
		}
		fmt.Fprintf(b, "BARCODE %d,%d,%q,%d,%d,%d,%d,%d,%q\n",
			d.X.At(dpi), d.Y.At(dpi), strings.ToUpper(d.Code.Name), h,
			d.Code.Wide.At(dpi), rot, d.Code.Human, d.Align, d.Data)
	case "qrcode":
		fmt.Fprintf(b, "QRCODE %d,%d,%s,%d,A,%d,M2,S7,%q\n",
			d.X.At(dpi), d.Y.At(dpi), strings.ToUpper(d.Code.Name),
			d.Code.Wide.At(dpi), rot, d.Data)
	default:
		return fmt.Errorf("layout %s not supported", d.Kind)
	}
	return nil
}

func fontSize(n *layla.Node) (res int) {
	if n.Font != nil {
		res = int(n.Font.Size)
	}
	if res == 0 {
		res = 8
	}
	return res
}
