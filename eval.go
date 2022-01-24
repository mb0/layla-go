package layla

import (
	"fmt"
	"io"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/ext"
	"xelf.org/xelf/lib"
	"xelf.org/xelf/lit"
)

// Eval parses and evaluates the label from reader r and returns a node or an error.
func Eval(ctx exp.Ctx, reg *lit.Reg, env exp.Env, rr io.Reader, name string) (*Node, error) {
	x, err := exp.Read(reg, rr, name)
	if err != nil {
		return nil, err
	}
	r, err := exp.EvalExp(ctx, reg, env, x)
	if err != nil {
		return nil, err
	}
	n := ValNode(r.Val)
	if n == nil {
		return nil, fmt.Errorf("expected *layla.Node got %T", r)
	}
	return n, nil
}

func ValNode(v lit.Val) *Node {
	if prx, ok := v.Value().(lit.Mut); ok {
		if n, ok := prx.Ptr().(*Node); ok {
			return n
		}
	}
	return nil
}

var listNodes = []string{"stage", "rect", "ellipse", "box", "vbox", "hbox", "table",
	"page", "extra", "cover", "header", "footer"}
var dataNodes = []string{"line", "text", "markup", "qrcode", "barcode"}

func Specs(reg *lit.Reg) lib.Specs {
	specs := make(lib.Specs, len(listNodes)+len(dataNodes))
	for _, name := range listNodes {
		s, err := ext.NodeSpecName(reg, name, &Node{Kind: name}, ext.Rules{Tail: ext.Rule{
			Prepper: ext.ListPrepper,
			Setter: func(p *exp.Prog, n ext.Node, _ string, v lit.Val) error {
				o := ValNode(n)
				for _, e := range v.(*lit.List).Vals {
					if e.Zero() {
						continue
					}
					c := ValNode(e)
					if c == nil {
						return fmt.Errorf("not a layla node %T", e)
					}
					if c.Align == 0 {
						c.Align = o.Align
					}
					if c.Font == nil {
						c.Font = o.Font
					}
					o.List = append(o.List, c)
				}
				return n.SetKey("list", v)
			},
		}})
		if err != nil {
			panic(err)
		}
		specs[name] = s
	}
	for _, name := range dataNodes {
		s, err := ext.NodeSpecName(reg, name, &Node{Kind: name}, ext.Rules{Tail: ext.Rule{
			Prepper: ext.DynPrepper,
			Setter: func(p *exp.Prog, n ext.Node, _ string, v lit.Val) error {
				return n.SetKey("data", v)
			},
		}})
		if err != nil {
			panic(err)
		}
		specs[name] = s
	}
	return specs
}
