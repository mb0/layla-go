package layla

import (
	"bytes"
	"strings"

	"xelf.org/layla/font"
	"xelf.org/layla/mark"
	"xelf.org/xelf/cor"
)

func (l *Layouter) lineLayout(n *Node, stack []*Node) (err error) {
	markup := n.Kind == "markup"
	var els []mark.El
	if markup {
		els, err = mark.Inline(n.Data)
		if err != nil {
			return err
		}
	} else {
		els = []mark.El{{Cont: n.Data}}
	}
	stack = append(stack, n)
	of := getFont(stack)
	b := n.Pad.Inset(n.Calc)
	lh, err := l.lineHeight(of)
	if err != nil {
		return err
	}
	s := &splitter{Layouter: l, Font: *of, Max: b.W}
	res, err := s.lines(els)
	if err != nil {
		return err
	}
	if markup {
		n.List = make([]*Node, 0, len(res))
	}
	var buf bytes.Buffer
	var y, mw Dot
	for li, line := range res {
		bx := b.X
		switch n.Align {
		case 2: // center
			bx += ((b.W - line.W) / 2).Floor()
		case 3: // right
			bx += (b.W - line.W).Floor()
		}
		if !markup && li > 0 {
			buf.WriteByte('\n')
		}
		var x Dot
		for _, sp := range line.Spans {
			w := sp.W
			if !markup {
				buf.WriteString(sp.Text)
			} else if sp.Text != " " {
				of := of
				if sp.Tag != 0 {
					ofv := *of
					ofv.Style = sp.Tag
					of = &ofv
				}
				n.List = append(n.List, &Node{
					Kind: "text",
					Data: sp.Text,
					Calc: Box{
						Pos: Pos{X: bx + x, Y: b.Y + y},
						Dim: Dim{W: w, H: lh},
					},
					Font: of,
				})
			}
			if x+w > mw {
				mw = x + w
			}
			x += w
		}
		y += lh
	}
	if !markup {
		n.Data = buf.String()
	}
	b.H = y
	b.W = mw
	b = n.Pad.Outset(b)
	n.Calc.H = clamp(n.Calc.H, b.H)
	if n.W > 0 {
		n.Calc.W = clamp(n.Calc.W, n.W)
	} else {
		n.Calc.W = clamp(n.Calc.W, b.W)
	}
	n.Font = of
	return nil
}

func (l *Layouter) lineHeight(f *Font) (lh Dot, _ error) {
	ff, err := l.Styler(l.Manager, *f, mark.Text)
	if err != nil {
		return 0, err
	}
	f.Height = ff.Metrics().Height

	if f.Line <= 0 {
		f.Line = 1.2
	}
	if f.Line < 8 {
		f.Line = (f.Line * Dot(l.PtToDot(f.Height))).Ceil()
	}
	return f.Line, nil
}

type splitter struct {
	*Layouter
	Font
	Max Dot
}

func (s *splitter) lines(els []mark.El) (res []line, err error) {
	var cur line
	res = make([]line, 0, len(els)/8)
	for _, el := range els {
		// select face
		f, err := s.Styler(s.Manager, s.Font, el.Tag)
		if err != nil {
			return res, err
		}
		res, cur = s.spans(f, el.Tag, el.Cont, res, cur)
	}
	if len(cur.Spans) > 0 {
		res = append(res, cur)
	}
	return res, nil
}

type line struct {
	Spans []span
	W     Dot
}

type span struct {
	Text string
	W    Dot
	Tag  mark.Tag
}

func (s *splitter) splitSpan(f *font.Face, txt string, mw Dot) (w Dot, _, rest string) {
	res := Dot(f.Extra())
	last := rune(-1)
	for i, r := range txt {
		wr := Dot(f.Rune(r, last))
		if i > 0 && res+wr > mw {
			return res, txt[:i], txt[i:]
		}
		res += wr
		last = r
	}
	return res, txt, ""
}

func (s *splitter) spanW(f *font.Face, txt string) Dot {
	w, _ := f.Text(txt, -1)
	w += f.Extra()
	return Dot(w).Ceil()
}
func (s *splitter) spans(f *font.Face, tag mark.Tag, cont string, res []line, cur line) ([]line, line) {
	var space bool
	sdot := Dot(f.Rune(s.Spacer, -1)).Ceil()
	for _, txt := range toks(cont) {
		switch txt {
		case "":
			res = append(res, cur)
			cur = line{}
			space = false
			continue
		case " ":
			space = true
			continue
		}
		ww := s.spanW(f, txt)
		var ws Dot
		if space {
			ws = sdot
			space = false
		}
		mw := s.Max - cur.W
		if ww+ws < mw { // normal case: fits in cur line
			if ws > 0 {
				cur.Spans = append(cur.Spans, span{" ", ws, tag})
			}
			cur.Spans = append(cur.Spans, span{txt, ww, tag})
			cur.W += ws + ww
			continue
		}
		// check for soft break point
		if d := strings.IndexRune(txt, '-'); d > 0 {
			fst, snd := txt[:d+1], txt[d+1:]
			wf := s.spanW(f, fst)
			if ws+wf < mw {
				if ws > 0 {
					cur.Spans = append(cur.Spans, span{" ", ws, tag})
				}
				cur.Spans = append(cur.Spans, span{fst, wf, tag})
				cur.W += ws + wf
				ww, ws = s.spanW(f, snd), 0
				txt = snd
			}
		}
		// we need to break the line
		// if the span does not fit the new line break inside the word until it does
		if ww > s.Max {
			i := 0
			for mw := s.Max - cur.W; ws+ww > mw; mw = s.Max {
				if i > 0 {
					if len(cur.Spans) > 0 {
						res = append(res, cur)
					}
					cur = line{}
				}
				cw, ct, rest := s.splitSpan(f, txt, mw-ws)
				cur.W += ws + cw
				if ws > 0 {
					cur.Spans = append(cur.Spans, span{" ", ws, tag})
					ws = 0
				}
				cur.Spans = append(cur.Spans, span{ct, cw, tag})
				ww = s.spanW(f, rest)
				txt = rest
				i++
			}
		}
		if len(cur.Spans) > 0 {
			res = append(res, cur)
		}
		cur = line{W: ww, Spans: []span{{txt, ww, tag}}}
	}
	if space {
		cur.Spans = append(cur.Spans, span{" ", sdot, tag})
		cur.W += sdot
	}
	return res, cur
}

func toks(text string) (res []string) {
	var start int
	var space bool
	for i, c := range text {
		if c == '\n' {
			if !space && i > start {
				res = append(res, text[start:i])
			}
			if space && len(res) > 0 && res[len(res)-1] == " " {
				res[len(res)-1] = ""
			} else {
				res = append(res, "")
			}
			space = true
			start = i + 1
		} else if cor.Space(c) {
			if !space {
				if i > start {
					res = append(res, text[start:i])
				}
				res = append(res, " ")
				space = true
			}
			start = i + 1
		} else if i >= start {
			space = false
		}
	}
	if len(text) > start {
		res = append(res, text[start:])
	}
	return res
}
