package layla

import (
	"fmt"
	"strings"

	"xelf.org/layla/font"
)

func Page(n *Node) ([]*Node, error) {
	p := newPager(n)
	err := p.collect(n)
	if err != nil {
		return nil, err
	}
	var res []*Node
	total := fmt.Sprint(len(p.list))
	for i, x := range p.list {
		if i > 0 {
			res = append(res, &Node{Kind: "page"})
		}
		x.page = fmt.Sprint(i + 1)
		x.total = total
		if p.Extra != nil {
			res = x.collect(p.Extra, res, 0)
		}
		top := p.Cover
		if i > 0 {
			top = p.Header
		}
		if top != nil {
			res = x.collect(top, res, 0)
		}
		if p.Footer != nil {
			offy := x.Y + x.H
			res = x.collect(p.Footer, res, offy)
		}
		res = append(res, x.res...)
	}
	return res, nil
}

type page struct {
	Org Dot
	Box
	res   []*Node
	page  string
	total string
}

func collectCopy(n *Node) *Node {
	d := &Node{Kind: n.Kind, Box: n.Calc, Border: n.Border}
	d.Pad = n.Pad
	switch n.Kind {
	case "line":
		d.Cols = append(([]font.Dot)(nil), n.Cols...)
	case "text":
		d.Font = n.Font
		d.Data = n.Data
		d.Align = n.Align
		d.Mar = n.Mar
	case "qrcode", "barcode":
		d.Code = n.Code
		d.Data = n.Data
	}
	return d
}

func (x *page) collect(n *Node, res []*Node, offy Dot) []*Node {
	var d *Node
	switch n.Kind {
	case "text":
		d = collectCopy(n)
		d.Data = strings.ReplaceAll(d.Data, "µP", x.page)
		d.Data = strings.ReplaceAll(d.Data, "µT", x.total)
	case "line", "qrcode", "barcode":
		d = collectCopy(n)
	case "rect", "ellipse":
		d = collectCopy(n)
		d.Y += offy
		res = append(res, d)
		fallthrough
	case "stage", "box", "vbox", "hbox", "table", "page",
		"extra", "cover", "header", "footer", "markup":
		for _, e := range n.List {
			res = x.collect(e, res, offy)
		}
		return res
	}
	d.Y += offy
	return append(res, d)
}

type pager struct {
	*Node
	Extra  *Node
	Cover  *Node
	Header *Node
	Footer *Node
	THead  []*Node
	list   []*page
}

func newPager(n *Node) *pager {
	p := &pager{Node: n}
	for _, e := range n.List {
		switch e.Kind {
		case "extra":
			p.Extra = e
		case "cover":
			p.Cover = e
		case "header":
			p.Header = e
		case "footer":
			p.Footer = e
		}
	}
	p.newPage(0)
	return p
}

func (p *pager) newPage(org Dot) *page {
	b := p.Pad.Inset(Box{Dim: p.Dim})
	top := p.Cover
	if len(p.list) > 0 {
		top = p.Header
	}
	if top != nil {
		h := top.Calc.H
		b.Y += h
		b.H -= b.Y
	}
	if p.Footer != nil {
		b.H -= p.Footer.Calc.H
	}
	x := &page{Org: org, Box: b}
	var mh Dot
	for _, th := range p.THead {
		if th.Calc.H > mh {
			mh = th.Calc.H
		}
		x.res = x.collect(th, x.res, x.Y-th.Calc.Y)
	}
	if mh > 0 {
		x.Y += mh
		x.H -= mh
	}
	p.list = append(p.list, x)
	return x
}

func (p *pager) collect(n *Node) error {
	switch n.Kind {
	case "text", "line", "qrcode", "barcode":
		p.draw(collectCopy(n), n.Mar)
	case "rect", "ellipse":
		p.draw(collectCopy(n), n.Mar)
		return p.collectAll(n.List)
	case "table":
		if p.Kind == "page" && n.Nobr && !p.fits(n) {
			p.newPage(n.Calc.Y)
		}
		hh := n.Head && len(p.THead) == 0
		if hh {
			head := n.List
			if len(head) > len(n.Cols) {
				head = head[:len(n.Cols)]
			}
			p.THead = head
		}
		err := p.collectAll(n.List)
		if hh {
			p.THead = nil
		}
		return err
	case "stage", "box", "vbox", "hbox", "page", "markup":
		return p.collectAll(n.List)
	case "extra", "cover", "header", "footer":
	}
	return nil
}

func (p *pager) collectAll(ns []*Node) (err error) {
	for _, e := range ns {
		err := p.collect(e)
		if err != nil {
			return err
		}
	}
	return nil
}
func (p *pager) fits(n *Node) bool {
	// check if case fits into the remaining space
	for i := len(p.list) - 1; i >= 0; i-- {
		x := p.list[i]
		if x.Org > n.Calc.Y {
			continue
		}
		y := n.Calc.Y - x.Org
		return y+n.Calc.H <= x.H
	}
	return false
}

func (p *pager) draw(n *Node, m *Off) {
	if p.Kind != "page" {
		xp := p.list[0]
		xp.res = append(xp.res, n)
		return
	}
	// find starting page
	for i := len(p.list) - 1; i >= 0; i-- {
		x := p.list[i]
		if x.Org > n.Y {
			continue
		}
		y := n.Y - x.Org
		// simple case fits into the remaining space
		if y+n.H <= x.H {
			n.Y = x.Y + y
			x.res = append(x.res, n)
			return
		}
		switch n.Kind {
		case "text":
			txt := strings.Split(n.Data, "\n")
			lh := n.Font.Line
			ah := x.H - y
			var hh Dot
			for j := 0; len(txt) > 0; j++ {
				lc := int(ah / lh)
				if lc == 0 && j > 0 {
					lc = 1
				}
				if lc > len(txt) {
					lc = len(txt)
				}
				if lc > 0 {
					nn := *n
					nn.Y = x.Y + y
					var padh Dot
					if nn.Pad != nil {
						padh = nn.Pad.T + nn.Pad.B
					}
					nn.H = (lh*Dot(lc) + padh).Ceil()
					hh += nn.H
					nn.Data = strings.Join(txt[:lc], "\n")
					x.res = append(x.res, &nn)
					txt = txt[lc:]
				}
				if len(txt) == 0 {
					return
				}
				i++
				if i < len(p.list) {
					x = p.list[i]
				} else {
					x = p.newPage(n.Y + hh)
				}
				y = 0
				if m != nil {
					y = m.T
				}
				ah = x.H
			}
			return
		}
		i++
		if i < len(p.list) {
			x = p.list[i]
		} else {
			x = p.newPage(n.Y)
		}
		n.Y = x.Y
		if m != nil {
			n.Y += m.T
		}
		x.res = append(x.res, n)
		return
	}
}
