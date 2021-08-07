package font

import (
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

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
	Add float64
}

func (f *Face) Extra() float64 { return f.Add }

func (f *Face) Text(text string, last rune) (res float64, _ rune) {
	for _, r := range text {
		a := f.Rune(r, last)
		res += a
		last = r
	}
	return res, last
}

func (f *Face) Rune(r, last rune) float64 {
	var res Pt
	if last != -1 && last != '\n' {
		res += f.Kern(last, r)
	}
	a, ok := f.GlyphAdvance(r)
	if !ok {
		a, _ = f.GlyphAdvance('x')
	}
	res += a
	d := f.PtToDot(res)
	switch f.subx {
	case 1:
		return math.Floor(d)
	case 2, 0:
		return math.Floor(d*2) / 2
	}
	subx := float64(f.subx)
	return math.Floor(d*subx) / subx
}
