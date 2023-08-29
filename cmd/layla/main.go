package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"xelf.org/layla"
	"xelf.org/layla/font"
	"xelf.org/layla/html"
	"xelf.org/layla/pdf"
	"xelf.org/layla/tsc"
	"xelf.org/layla/tspl"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lib/extlib"
	"xelf.org/xelf/lit"
)

var rend = flag.String("rend", "tspl", "renderer")
var fnt = flag.String("font", ".", "specific font file or directory (tspl only supports ttf)")
var dpi = flag.Int("dpi", 0, "resolution in dots per inch")
var prnt = flag.Int("print", 0, "number of labels to print")
var dev = flag.String("dev", "", "device string either dev path or net addr")

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("expect one or more arguments: template and optionally arguments dicts")
	}
	tmpl := args[0]
	if !strings.HasSuffix(tmpl, ".layla") {
		log.Fatal("expect template argument to have an .layla extension")
	}
	dp := *dpi
	if dp == 0 {
		dp = 203
	}
	man := font.NewManager(dp, 2, 2)
	err := registerFonts(man, *fnt)
	if err != nil {
		log.Println("read font: ", err)
	}
	err = registerFonts(man, "./testdata/font")
	if err != nil {
		log.Println("read test font: ", err)
	}
	var argmap lit.Dict
	argmap.SetKey("now", lit.Time(time.Now()))
	for _, arg := range args[1:] {
		cb, err := ioutil.ReadFile(arg)
		if err != nil {
			log.Fatalf("read ctx: %v", err)
		}
		ctx, err := lit.Read(bytes.NewReader(cb), arg)
		if err != nil {
			log.Fatalf("parse ctx: %v", err)
		}
		keyr, ok := ctx.(lit.Keyr)
		if !ok {
			log.Fatalf("expect keyr got %T", ctx)
		}
		err = keyr.IterKey(func(k string, v lit.Val) error {
			return argmap.SetKey(k, v)
		})
		if err != nil {
			log.Fatalf("update arg map: %v", err)
		}

	}
	tb, err := ioutil.ReadFile(tmpl)
	if err != nil {
		log.Fatal("read tmpl: ", err)
	}
	x, err := exp.Read(bytes.NewReader(tb), tmpl)
	if err != nil {
		log.Fatal("parse tmpl: ", err)
	}
	env := exp.Builtins(layla.Specs(nil).AddMap(extlib.Std))
	r, err := exp.NewProg(env).Run(x, &argmap)
	if err != nil {
		log.Fatal("run tmpl: ", err)
	}
	node := layla.ValNode(r)
	if node == nil {
		log.Fatal("exec tmpl: ", err)
	}
	name := filepath.Base(tmpl)
	out := filepath.Join(filepath.Dir(tmpl), name[:len(name)-6])
	var buf bytes.Buffer
	switch *rend {
	case "tspl":
		pre := ""
		if *prnt != 0 {
			pre = "SET KEY1 PRINT 1\nDENSITY 15"
		}
		err = tspl.Render(&buf, man, node, pre)
		if err != nil {
			log.Fatal("render: ", err)
		}
		if *prnt == 0 {
			break
		}
		fmt.Fprintf(&buf, "PRINT %d\n", *prnt)
		c, err := tsc.Auto(*dev, time.Second)
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()
		_, err = io.Copy(c.Conn, &buf)
		if err != nil {
			log.Printf("could not copy: %v", err)
		}
		raw, _ := c.Recv()
		if len(raw) > 0 {
			log.Printf("received %d bytes:\n%s", len(raw), raw)
		}
		return
	case "html":
		var b bytes.Buffer
		b.WriteString("<body style=\"background-color: grey\">\n")
		err = html.Render(&b, man, node)
		if err != nil {
			log.Fatalf("render html error: %v", err)
		}
		b.WriteString(`</body>`)
		err = ioutil.WriteFile(out+".html", b.Bytes(), 0644)
		if err != nil {
			log.Printf("write html error: %v", err)
		}
	case "pdf":
		doc, err := pdf.Render(man, node)
		if err != nil {
			log.Fatalf("render %q error: %v", name, err)
		}
		err = doc.OutputFileAndClose(out + ".pdf")
		if err != nil {
			log.Printf("write error: %v", err)
		}

	default:
		log.Fatalf("expect format argument to be either of tspl, html or pdf got %s", *rend)
	}
	fmt.Print(buf.String())
}

func registerFonts(man *font.Manager, path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("could not stat font path: %v", err)
	}
	var files []string
	if fi.IsDir() {
		fis, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("could not read fonts: %v", err)
		}
		for _, fi := range fis {
			name := fi.Name()
			if isFontPath(name) {
				files = append(files, filepath.Join(path, name))
			}
		}
	} else if isFontPath(path) {
		files = append(files, path)
	}
	for _, file := range files {
		// log.Printf("register font %s", file)
		man.Register(filepath.Base(file), file)
	}
	return man.Err()
}

func isFontPath(path string) bool {
	if len(path) > 4 {
		switch strings.ToLower(path[len(path)-4:]) {
		case ".ttf", ".otf":
			return true
		}
	}
	return false
}
