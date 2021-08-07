package font

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

type Key struct {
	Name string
	Size float64
}

type Src struct {
	*truetype.Font
	Path string
}

type Manager struct {
	dpi   float64
	subx  int
	suby  int
	ttfs  map[string]*Src
	faces map[Key]font.Face
	err   error
}

func NewManager(dpi, subx, suby int) *Manager {
	return &Manager{dpi: float64(dpi), subx: subx, suby: suby}
}

func (m *Manager) DPI() float64 {
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

func (m *Manager) DotToPt(dot float64) Pt { return PtF(dot * m.DPI() / (25.4 * 8)) }
func (m *Manager) PtToDot(pt Pt) float64  { return PtToF(pt) * 25.4 * 8 / m.DPI() }

func (m *Manager) Err() error           { return m.err }
func (m *Manager) fail(err error) error { m.err = err; return err }

func (m *Manager) RegisterTTF(name string, path string) *Manager {
	if m.err != nil {
		return m
	}
	_, ok := m.ttfs[name]
	if ok {
		return m
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		m.fail(fmt.Errorf("reading file %q: %v", path, err))
		return m
	}
	f, err := truetype.Parse(data)
	if err != nil {
		m.fail(fmt.Errorf("parse file %q: %v", path, err))
		return m
	}
	if m.ttfs == nil {
		m.ttfs = make(map[string]*Src)
	}
	m.ttfs[name] = &Src{f, path}
	return m
}

func (m *Manager) Path(name string) (string, error) {
	src, ok := m.ttfs[name]
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
	src, ok := m.ttfs[name]
	if !ok {
		return nil, fmt.Errorf("unknown font %q", name)
	}
	subx, suby := m.SubPixels()
	f = truetype.NewFace(src.Font, &truetype.Options{
		Size:       size,
		DPI:        m.DPI(),
		SubPixelsX: subx,
		SubPixelsY: suby,
	})
	if m.faces == nil {
		m.faces = make(map[Key]font.Face)
	}
	m.faces[key] = f
	return f, nil
}
