package font

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Key struct {
	Name string
	Size float64
}

type Src struct {
	Font interface{}
	Path string
	Name string
}

type Manager struct {
	Compat bool
	dpi    int
	subx   int
	suby   int
	srcs   map[string]*Src
	faces  map[Key]font.Face
	err    error
}

func NewManager(dpi, subx, suby int) *Manager {
	return &Manager{dpi: dpi, subx: subx, suby: suby}
}

func (m *Manager) DPI() int {
	if m.dpi <= 0 {
		return 72
	}
	return m.dpi
}

func (m *Manager) SubPixels() (x, y int) {
	if x = m.subx; x <= 0 {
		x = 2
	}
	if y = m.suby; y <= 0 {
		y = 4
	}
	return x, y
}

func (m *Manager) DotToPt(dot Dot) Pt { return PtF(float64(dot * Dot(m.DPI()) / (25.4 * 8))) }
func (m *Manager) PtToDot(pt Pt) Dot  { return Dot(PtToF(pt)*25.4*8) / Dot(m.DPI()) }

func (m *Manager) Err() error           { return m.err }
func (m *Manager) fail(err error) error { m.err = err; return err }

// Regster supports both truetype and opentype fonts.
// TSC printers however only support truetype fonts. Opentype fonts can still be used
// to generate pdf documents.
func (m *Manager) Register(name string, path string) *Manager {
	if m.err != nil {
		return m
	}
	_, ok := m.srcs[name]
	if ok {
		return m
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		m.fail(fmt.Errorf("reading file %q: %v", path, err))
		return m
	}
	var f interface{}
	switch strings.ToLower(path[len(path)-4:]) {
	case ".otf":
		f, err = opentype.Parse(data)
	case ".ttf":
		f, err = truetype.Parse(data)
	default:
		err = fmt.Errorf("unknown font format")
	}
	if err != nil {
		m.fail(fmt.Errorf("parse file %q: %v", path, err))
		return m
	}
	if m.srcs == nil {
		m.srcs = make(map[string]*Src)
	}
	m.srcs[name] = &Src{f, path, name}
	return m
}

func (m *Manager) Path(name string) (string, error) {
	src, ok := m.srcs[name]
	if !ok {
		return "", fmt.Errorf("unknown font %q", name)
	}
	return src.Path, nil
}

func (m *Manager) Face(name string, size float64) (font.Face, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := Key{name, size}
	f, ok := m.faces[key]
	if ok {
		return f, nil
	}
	src, ok := m.srcs[name]
	if !ok {
		return nil, fmt.Errorf("unknown font %q", name)
	}
	subx, suby := m.SubPixels()
	switch fnt := src.Font.(type) {
	case *truetype.Font:
		f = truetype.NewFace(fnt, &truetype.Options{
			Size:       size,
			DPI:        float64(m.DPI()),
			SubPixelsX: subx,
			SubPixelsY: suby,
		})
	case *opentype.Font:
		f, _ = opentype.NewFace(fnt, &opentype.FaceOptions{
			Size: size,
			DPI:  float64(m.DPI()),
		})
	}
	if m.faces == nil {
		m.faces = make(map[Key]font.Face)
	}
	m.faces[key] = f
	return f, nil
}
