package layla

import (
	"reflect"
	"testing"

	"xelf.org/layla/font"
)

func TestLayout(t *testing.T) {
	m := font.NewManager(72, 2, 4).RegisterTTF("", "testdata/font/Go-Regular.ttf")
	tests := []struct {
		text  string
		width int
		want  string
	}{
		{"Hello world", 66, "Hello world"},
		{"Hello world", 33, "Hello\nworld"},
		{"Hello world", 22, "Hell\no w\norld"},
		{"Tobeornottobe", 22, "Tob\neor\nnott\nobe"},
		{"To be or not to be", 50, "To be or\nnot to be"},
		{"To be or_not to be", 50, "To be\nor_not to\nbe"},
		{"To be or-not to be", 54, "To be or-\nnot to be"},
		{"To be\nor not\nto be", 50, "To be\nor not\nto be"},
	}
	lay := &Layouter{m, ' ', ZeroStyler}
	for i, test := range tests {
		n := &Node{
			Kind: "text",
			Data: test.text,
			Calc: Box{Dim: Dim{W: Dot(m.PtToDot(font.PtI(test.width)))}},
		}
		err := lay.lineLayout(n, nil)
		if err != nil {
			t.Errorf("layout error: %v", err)
			continue
		}
		if !reflect.DeepEqual(test.want, n.Data) {
			t.Errorf("test %d want lines %q got %q", i, test.want, n.Data)
		}
	}
}

func TestTokens(t *testing.T) {
	tests := []struct {
		text string
		want []string
	}{
		{"Mr. A BC", []string{"Mr.", " ", "A", " ", "BC"}},
		{"x y", []string{"x", " ", "y"}},
		{"x  \n  y", []string{"x", "", "y"}},
		{"x  \ny", []string{"x", "", "y"}},
		{"x  \n  \n  y", []string{"x", "", "", "y"}},
		{"foo  bar", []string{"foo", " ", "bar"}},
		{"-o-  bar", []string{"-o-", " ", "bar"}},
		{"foo-bar", []string{"foo-bar"}},
	}
	for _, test := range tests {
		got := toks(test.text)
		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("test %q want toks %v got %v", test.text, test.want, got)
		}
	}
}
