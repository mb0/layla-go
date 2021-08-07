package mark

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		raw string
		res []El
	}{
		{"test", []El{
			{Tag: Para, Els: []El{{Cont: "test"}}},
		}},
		{"test\ntest", []El{
			{Tag: Para, Els: []El{{Cont: "test"}, {Cont: "test"}}},
		}},
		{"test\n\ntest", []El{
			{Tag: Para, Els: []El{{Cont: "test"}}},
			{Tag: Para, Els: []El{{Cont: "test"}}},
		}},
		{"test\n------\ntest", []El{
			{Tag: Para, Els: []El{{Cont: "test"}}},
			{Tag: Ruler},
			{Tag: Para, Els: []El{{Cont: "test"}}},
		}},
		{"#title\n###title\n########title\ntest", []El{
			{Tag: Head1, Cont: "title"},
			{Tag: Head3, Cont: "title"},
			{Tag: Head4, Cont: "title"},
			{Tag: Para, Els: []El{{Cont: "test"}}},
		}},
		{"test [Link](url) *test* _test_ `  test   ` test", []El{
			{Tag: Para, Els: []El{
				{Cont: "test "},
				{Tag: Link, Cont: "url", Els: []El{{Cont: "Link"}}},
				{Cont: " "},
				{Tag: Bold, Cont: "test"},
				{Cont: " "},
				{Tag: Italic, Cont: "test"},
				{Cont: " "},
				{Tag: Code, Cont: "  test   "},
				{Cont: " test"},
			}},
		}},
	}
	for _, test := range tests {
		res, err := Parse(test.raw)
		if err != nil {
			t.Errorf("parse %q err: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(res, test.res) {
			t.Errorf("parse %q\nwant %v\n got %v", test.raw, test.res, res)
		}
	}
}
