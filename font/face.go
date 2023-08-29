package font

import (
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Dot is the measurement used by layla and is defined as 1/8 mm or 1/203 inch.
// We use the 200 dpi dot as default, because it easier to work with than mm or 300 dpi dots.
// During layout we round to half dots so we do not lose information relevant for 300 dpi renders.
// Otherwise we try to round positions down and dimensions up, as not to squeeze and wrap text.
type Dot float64

// Mm converts dot into mm and returns it.
func (dot Dot) Mm() float64 { return float64(dot / 8) }

func (dot Dot) Round() Dot     { return Dot(math.Round(float64(dot))) }
func (dot Dot) RoundHalf() Dot { return Dot(math.Round(float64(dot*2))) / 2 }

func (dot Dot) Ceil() Dot     { return Dot(math.Ceil(float64(dot))) }
func (dot Dot) CeilHalf() Dot { return Dot(math.Ceil(float64(dot*2))) / 2 }

func (dot Dot) Floor() Dot     { return Dot(math.Floor(float64(dot))) }
func (dot Dot) FloorHalf() Dot { return Dot(math.Floor(float64(dot*2))) / 2 }

// At returns dot at a specific resolution in dots per inch.
func (dot Dot) At(dpi int) int {
	if dpi < 200 || dpi > 203 {
		dot = dot * Dot(dpi) / 203
	}
	return int(dot.Round())
}

type Pt = fixed.Int26_6

var PtI = fixed.I

func PtF(f float64) Pt {
	return Pt(f * 64)
}

func PtToF(pt Pt) float64 {
	res := float64(pt >> 6)
	res += float64(pt&63) / 64
	return res
}

type Face struct {
	*Manager
	font.Face
	Add Dot
}

func (f *Face) Extra() Dot { return f.Add }

func (f *Face) Text(text string, last rune) (res Dot, _ rune) {
	for _, r := range text {
		a := f.Rune(r, last)
		res += a
		last = r
	}
	return res, last
}

func (f *Face) Rune(r, last rune) Dot {
	var res Pt
	if last != -1 && last != '\n' && last != ' ' {
		res += f.Kern(last, r)
	}
	a, ok := f.GlyphAdvance(r)
	if !ok {
		a, _ = f.GlyphAdvance('X')
	}
	res += a
	d := f.PtToDot(res)
	if f.subx == 1 {
		return d.Floor()
	}
	return d.FloorHalf()
}
