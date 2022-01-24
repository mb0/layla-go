package layla_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"xelf.org/layla"
	"xelf.org/layla/font"
	"xelf.org/layla/html"
	"xelf.org/layla/pdf"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lib/extlib"
	"xelf.org/xelf/lit"
)

func man() *font.Manager {
	m := font.NewManager(72, 2, 4).
		RegisterTTF("GoReg.ttf", "testdata/font/Go-Regular.ttf").
		RegisterTTF("GoBold.ttf", "testdata/font/Go-Bold.ttf")
	if err := m.Err(); err != nil {
		log.Fatal(err)
	}
	// enabled tspl text rendering compatibility mode
	m.Compat = true
	return m
}

var testFiles = []string{
	"test",
	"lines",
	"textbox",
	"pages",
	"label1",
	"label2",
}

func TestHtml(t *testing.T) {
	m := man()
	for _, name := range testFiles {
		n, err := read(exp.BG, name)
		if err != nil {
			t.Errorf("error reading test file %q: %v", name, err)
			continue
		}
		var b bytes.Buffer
		b.WriteString("<body style=\"background-color: grey\">\n")
		err = html.Render(&b, m, n)
		if err != nil {
			t.Errorf("render html error: %v", err)
			continue
		}
		b.WriteString(`</body>`)
		err = ioutil.WriteFile(path(name, ".html"), b.Bytes(), 0644)
		if err != nil {
			t.Errorf("write html error: %v", err)
		}
	}
}

func TestPdf(t *testing.T) {
	m := man()
	for _, name := range testFiles {
		n, err := read(exp.BG, name)
		if err != nil {
			t.Errorf("error reading test file %q: %v", name, err)
			continue
		}
		doc, err := pdf.Render(m, n)
		if err != nil {
			t.Errorf("render %q error: %v", name, err)
			continue
		}
		err = doc.OutputFileAndClose(path(name, ".pdf"))
		if err != nil {
			t.Errorf("write error: %v", err)
		}
	}
}

func read(ctx exp.Ctx, name string) (*layla.Node, error) {
	f, err := os.Open(path(name, ".layla"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	now := time.Date(2019, time.October, 5, 23, 0, 0, 0, time.UTC)
	param := &lit.Dict{Keyed: []lit.KeyVal{
		{"now", lit.Time(now)},
		{"title", lit.Str("Produkt")},
		{"vendor", lit.Str("Firma GmbH")},
		{"batch", lit.Str("AB19020501")},
		{"ingreds", lit.Str("list of all the ingredients, like suger and spice and everthing nice.")},
	}}
	reg := &lit.Reg{}
	env := exp.Builtins(layla.Specs(reg).AddMap(extlib.Std))
	return layla.Eval(ctx, reg, &exp.ArgEnv{Par: env, Typ: param.Type(), Val: param}, f, name+".layla")
}

func path(name, ext string) string {
	return filepath.Join("testdata", name+ext)
}
