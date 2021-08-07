package mark

import (
	"strings"

	"xelf.org/xelf/cor"
)

type Tag uint

const (
	Bold Tag = 1 << iota
	Italic
	Code
	Link

	Head1
	Head2
	Head3
	Head4
	Ruler
	Para

	Text   = Tag(0)
	Style  = Bold | Italic | Code | Link
	Header = Head1 | Head2 | Head3 | Head4
	Block  = Ruler | Para
	All    = Style | Header | Block
)

type El struct {
	Tag  Tag
	Cont string
	Els  []El
}

func Parse(txt string) ([]El, error)  { return All.Parse(txt) }
func Inline(txt string) ([]El, error) { return All.Inline(txt) }

func (tag Tag) Parse(txt string) (res []El, err error) {
	var line string
	var cont bool
	for len(txt) > 0 {
		line, txt = readLine(txt)
		var el *El
		switch {
		case tag&Header != 0 && strings.HasPrefix(line, "#"):
			var c int
			for c = 1; c < len(line); c++ {
				if line[c] != '#' {
					break
				}
			}
			el = &El{Cont: strings.TrimSpace(line[c:])}
			if c > 4 {
				c = 4
			}
			el.Tag = Head1 << (c - 1)
			cont = false
		case tag&Ruler != 0 && strings.HasPrefix(line, "---"):
			el = &El{Tag: Ruler}
			line = strings.TrimLeft(line, "-")
			el.Cont = strings.TrimSpace(line)
			cont = false
		default:
			line = strings.TrimSpace(line)
			if line == "" {
				cont = false
				continue
			}
			els, err := tag.Inline(line)
			if err != nil {
				return res, err
			}
			if cont {
				el = &res[len(res)-1]
				el.Els = append(el.Els, els...)
				continue
			}
			el = &El{Tag: Para, Els: els}
			cont = true
		}
		res = append(res, *el)
	}
	return
}

func readLine(txt string) (line, rest string) {
	end := strings.IndexByte(txt, '\n')
	if end < 0 {
		return txt, ""
	}
	line, rest = txt[:end], txt[end+1:]
	if end > 0 && line[end-1] == '\r' {
		line = line[:end-1]
	}
	return
}

func (tag Tag) Inline(txt string) (res []El, _ error) {
	var start, i int
	for i < len(txt) {
		c := rune(txt[i])
		tag, end, ok := tag.inlineStart(c)
		switch ok {
		case true:
			cont, n := consumeSpan(txt[i:], end, tag == Code)
			if n == 0 {
				break
			}
			var link string
			var nn int
			if tag == Link {
				ii := i + n
				ii += skipSpace(txt[ii:])
				if ii >= len(txt) || txt[ii] != '(' {
					break
				}
				link, nn = consumeSpan(txt[ii:], ')', false)
				if nn == 0 {
					break
				}
				nn += ii - i - n
			}
			if start < i {
				cont := txt[start:i]
				res = append(res, El{Cont: cont})
			}
			i += n + nn
			el := El{Tag: tag, Cont: cont}
			if tag == Link {
				el.Els = []El{{Cont: el.Cont}}
				el.Cont = link
			}
			start = i
			res = append(res, el)
			continue

		}
		i++
		for _, c := range txt[i:] {
			if !cor.Space(c) {
				break
			}
			i++
		}
	}
	if start < len(txt) {
		res = append(res, El{Cont: txt[start:]})
	}
	return res, nil
}

func (tag Tag) inlineStart(c rune) (Tag, rune, bool) {
	switch {
	case tag&Link != 0 && c == '[': // link
		return Link, ']', true
	case tag&Bold != 0 && c == '*': // emphasis
		return Bold, c, true
	case tag&Italic != 0 && c == '_':
		return Italic, c, true
	case tag&Code != 0 && c == '`':
		return Code, c, true
	}
	return Text, 0, false
}

func skipSpace(s string) (n int) {
	for _, c := range s {
		if !cor.Space(c) {
			return
		}
		n++
	}
	return
}

func consumeSpan(txt string, end rune, spok bool) (string, int) {
	var esc bool
	for i, r := range txt {
		if i == 0 {
			continue
		}
		if !spok && i == 1 && r == ' ' {
			break
		}
		if esc {
			esc = false
			continue
		}
		switch r {
		case '\\':
			esc = true
		case end:
			return txt[1:i], i + 1
		}
	}
	return "", 0
}
