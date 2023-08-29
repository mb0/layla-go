// Package obpp defines a 1bit per pixel image format where 1 is white.
// The file format uses the 1bpp extension and has a 6 bytes header with the magic two bytes
// 0x31 0x62 ('1b') and then width and height each uint16 in big-endian followed by
// height*((width+pad)/8) bytes data, where pad is padding a row to full bytes (8-(width%8))%8.
//
// This code is not actively used or maintained.
//
// This code was written in anticipation of printing bitmap data with tsc printers. However, the
// targeted printer did not support BITMAP, PUTBMP or PUTPCX commands. Therefor I had to use
// the TTF renderer and a modified font file to print the logo.
package onebpp

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
)

func init() {
	image.RegisterFormat("1bpp", "1b", func(r io.Reader) (image.Image, error) {
		return Decode(r)
	}, func(r io.Reader) (image.Config, error) {
		b, err := readHeader(r)
		return image.Config{Width: b.W, Height: b.H, ColorModel: color.GrayModel}, err
	})
}

// Bitmap implements the image interface for a simple 1bit per pixel image format.
// The on disk format start with two magic bytes '1' and 'b', then width and height as big-endian
// uint16 and the bitmap data with rows padded to full bytes.
type Bitmap struct {
	W    int    `json:"w"`
	H    int    `json:"h"`
	Data []byte `json:"data"`
}

// Decode reads r and returns a 1bit per pixel bitmap image or an error.
func Decode(r io.Reader) (*Bitmap, error) {
	b, err := readHeader(r)
	if err != nil {
		return nil, err
	}
	pad := (8 - b.W%8) % 8
	b.Data = make([]byte, ((b.W+pad)/8)*b.H)
	_, err = io.ReadFull(r, b.Data)
	return b, err
}

// Convert returns a 1bit per pixel bitmap image using a default threshold and white padding.
func Convert(img image.Image) *Bitmap { return ConvertThreshold(img, 0x7f, true) }

// ConvertThreshold returns a 1bit per pixel bitmap using the 8bit grayscale threshold.
// If fill is true, any padding is painted white.
func ConvertThreshold(img image.Image, threshold uint8, fill bool) *Bitmap {
	if res, ok := img.(*Bitmap); ok {
		return res
	}
	b := img.Bounds()
	h := b.Max.Y - b.Min.Y
	iw := b.Max.X - b.Min.X
	// we may need to add some pixels to match the format
	pad := (8 - iw%8) % 8
	rw := (iw + pad) / 8
	// a zeroes byte slice is an all black image
	data := make([]byte, rw*h)
	var raw byte
	for y := 0; y < h; y++ {
		dof := y * rw
		x := 0
		for ; x < iw; x++ {
			c := img.At(x, y)
			cg, ok := c.(color.Gray)
			if !ok {
				r, g, b, _ := c.RGBA()
				cg.Y = uint8((19595*r + 38470*g + 7471*b + 1<<15) >> 24)
			}
			// and only paint white pixels
			if cg.Y > threshold {
				raw |= 1 << (7 - x%8)
			}
			if x != 0 && (x+1)%8 == 0 {
				data[dof+(x/8)] = raw
				raw = 0
			}
		}
		if pad > 0 {
			if fill { // fill with white
				raw |= 0xff >> (x % 8)
			}
			data[dof+(x-1)/8] = raw
			raw = 0
		}
	}
	return &Bitmap{W: iw, H: h, Data: data}
}

// following methods implement image.Image

func (*Bitmap) ColorModel() color.Model   { return color.GrayModel }
func (b *Bitmap) Bounds() image.Rectangle { return image.Rect(0, 0, b.W, b.H) }
func (b *Bitmap) At(x, y int) color.Color {
	pad := (8 - b.W%8) % 8
	o := y*((b.W+pad)/8) + (x % 8)
	if o >= 0 && o < len(b.Data) {
		if b.Data[o]&(1<<(7-(x%8))) == 0 {
			return color.Gray{Y: 0x00}
		}
	}
	return color.Gray{Y: 0xff}
}

// WriteTo writes the bitmap to w, returning number bytes written and any error.
func (b *Bitmap) WriteTo(w io.Writer) (int64, error) {
	var head [6]byte
	head[0], head[1] = '1', 'b'
	binary.BigEndian.PutUint16(head[2:], uint16(b.W))
	binary.BigEndian.PutUint16(head[4:], uint16(b.H))
	n, err := w.Write(head[:])
	if err != nil {
		return int64(n), err
	}
	n, err = w.Write(b.Data)
	nn := int64(n) + 6
	if err != nil {
		return nn, err
	}
	if n < len(b.Data) {
		return nn, io.ErrShortWrite
	}
	return nn, nil
}

func readHeader(r io.Reader) (*Bitmap, error) {
	var head [6]byte
	_, err := io.ReadFull(r, head[:])
	if err != nil {
		return nil, err
	}
	if head[0] != '1' && head[1] != 'b' {
		return nil, fmt.Errorf("invalid onebpp header")
	}
	w := binary.BigEndian.Uint16(head[2:])
	h := binary.BigEndian.Uint16(head[4:])
	return &Bitmap{W: int(w), H: int(h)}, nil
}
