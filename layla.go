package layla

import (
	"xelf.org/layla/font"
	"xelf.org/layla/mark"
)

const (
	AlignLeft = iota
	AlignRight
	AlignCenter
)

type Dot = font.Dot

// Pos is a simple position consisting of x and y coordinates in dots.
type Pos struct {
	X Dot `json:"x,omitempty"`
	Y Dot `json:"y,omitempty"`
}

// Dim is a simple dimension consisting of width and height in dots.
type Dim struct {
	W Dot `json:"w,omitempty"`
	H Dot `json:"h,omitempty"`
}

// Box is a simple box consisting of a position and dimension.
type Box struct {
	Pos
	Dim
}

// Off is a box offset consisting of left, top, right and bottom offsets in dot.
type Off struct {
	L Dot `json:"l,omitempty"`
	T Dot `json:"t,omitempty"`
	R Dot `json:"r,omitempty"`
	B Dot `json:"b,omitempty"`
}

// Inset returns a box result of b with o substracted.
func (o *Off) Inset(b Box) Box {
	if o != nil {
		b.X += o.L
		b.Y += o.T
		b.W -= o.L + o.R
		b.H -= o.T + o.B
		if b.W < 0 {
			b.W = 0
		}
		if b.H < 0 {
			b.H = 0
		}
	}
	return b
}

// Outset returns a box result of b with o added.
func (o *Off) Outset(b Box) Box {
	if o != nil {
		b.X -= o.L
		b.Y -= o.T
		b.W += o.L + o.R
		b.H += o.T + o.B
	}
	return b
}

// Font holds all font related node data
type Font struct {
	Name   string   `json:"name,omitempty"`
	Size   float64  `json:"size,omitempty"`
	Line   Dot      `json:"line,omitempty"`
	Style  mark.Tag `json:"-"`
	Height font.Pt  `json:"-"`
}

// NodeLayout holds all layout related node data
type NodeLayout struct {
	Mar   *Off `json:"mar,omitempty"`
	Pad   *Off `json:"pad,omitempty"`
	Rot   int  `json:"rot,omitempty"`
	Align int  `json:"align,omitempty"`
	Gap   Dot  `json:"gap,omitempty"`
	Sub   Dim  `json:"sub,omitempty"`
}

// Code holds all qr and barcode related node data
type Code struct {
	Name  string `json:"name,omitempty"`
	Human int    `json:"human,omitempty"`
	Wide  Dot    `json:"wide,omitempty"`
}

type Color struct {
	R int `json:"r,omitempty"`
	G int `json:"g,omitempty"`
	B int `json:"b,omitempty"`
}

type Border struct {
	W Dot `json:"w,omitempty"`
	L Dot `json:"l,omitempty"`
	T Dot `json:"t,omitempty"`
	R Dot `json:"r,omitempty"`
	B Dot `json:"b,omitempty"`
}

func (b Border) Default(w Dot) Border {
	if b.W <= 0 {
		b.W = w
	}
	if b.W > 0 && b.L == 0 && b.T == 0 && b.R == 0 && b.B == 0 {
		b.L = b.W
		b.T = b.W
		b.R = b.W
		b.B = b.W
	}
	return b
}

type Table struct {
	Cols []Dot `json:"cols,omitempty"`
	Head bool  `json:"head,omitempty"`
	Nobr bool  `json:"nobr,omitempty"`
}

// Node is a part of the display tree and can represent any element.
type Node struct {
	Kind string `json:"kind"`
	Box
	NodeLayout
	Font   *Font   `json:"font,omitempty"`
	Border Border  `json:"border,omitempty"`
	List   []*Node `json:"list,omitempty"`
	Table
	Code *Code  `json:"code,omitempty"`
	Data string `json:"data,omitempty"`
	Calc Box    `json:"-"`
}
