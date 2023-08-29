package layla

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/ext"
	"xelf.org/xelf/lib"
	"xelf.org/xelf/lit"
)

// Eval parses and evaluates the label from reader r and returns a node or an error.
func Eval(ctx context.Context, reg *lit.Regs, env exp.Env, rr io.Reader, name string) (*Node, error) {
	return EvalProg(exp.NewProg(env, reg, ctx), rr, name, nil)
}

// EvalProg parses and evaluates the label from reader using prog r and returns a node or an error.
func EvalProg(prog *exp.Prog, rr io.Reader, name string, arg interface{}) (*Node, error) {
	x, err := exp.Read(rr, name)
	if err != nil {
		return nil, err
	}
	var v lit.Val
	if arg != nil {
		v, err = prog.Reg.ProxyValue(reflect.ValueOf(arg))
		if err != nil {
			return nil, fmt.Errorf("proxy error: %v", err)
		}
	}
	r, err := prog.Run(x, v)
	if err != nil {
		return nil, err
	}
	n := ValNode(r)
	if n == nil {
		return nil, fmt.Errorf("expected *layla.Node got %T", r)
	}
	return n, nil
}

func ValNode(v lit.Val) *Node {
	if n, ok := v.Mut().Ptr().(*Node); ok {
		return n
	}
	return nil
}

var listNodes = []string{"stage", "rect", "ellipse", "box", "vbox", "hbox", "table",
	"page", "extra", "cover", "header", "footer"}
var dataNodes = []string{"line", "text", "markup", "qrcode", "barcode"}

func Specs(reg lit.Reg) lib.Specs {
	if reg == nil {
		reg = lit.GlobalRegs()
	}
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
			Prepper: ext.CatPrepper,
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
