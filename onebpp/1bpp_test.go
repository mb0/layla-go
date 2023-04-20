package onebpp

import (
	"bytes"
	"encoding/hex"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"reflect"
	"testing"
)

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

var (
	arrow = &Bitmap{W: 16, H: 16, Data: mustHex(
		`00000000000007FF03FF11FF18FF1C7F1E3F1F1F1F8F1FC71FE31FF71FFF1FFF`,
	)}
	chess = &Bitmap{W: 5, H: 4, Data: mustHex(`AF57AF57`)}
)

func TestBitmap(t *testing.T) {
	tests := []struct {
		name string
		bm   *Bitmap
		want []byte
	}{
		{"chess", chess, append([]byte{'1', 'b', 0, 5, 0, 4}, chess.Data...)},
	}
	for _, test := range tests {
		var buf bytes.Buffer
		test.bm.WriteTo(&buf)
		got := buf.Bytes()
		if !bytes.Equal(got, test.want) {
			t.Errorf("encode for %s\n got: %x \nwant: %x", test.name, got, test.want)
		}
		bm, _, err := image.Decode(bytes.NewReader(test.want))
		if err != nil {
			t.Errorf("decode %s failed: %v", test.name, err)
		}
		if !reflect.DeepEqual(bm, test.bm) {
			t.Errorf("decode for %s\n got: %+v \nwant: %+v", test.name, bm, test.bm)
		}
	}
}

func TestConvert(t *testing.T) {
	tests := []struct {
		file string
		want *Bitmap
	}{
		{"../testdata/arrow.bw.gif", arrow},
		{"../testdata/arrow.gray8.png", arrow},
		{"../testdata/arrow.rgb8.png", arrow},
		{"../testdata/arrow.rgba.jpg", arrow},
		{"../testdata/arrow.rgba16.png", arrow},
		{"../testdata/chess5x4.gif", chess},
	}
	for _, test := range tests {
		img, _, err := decImgFile(test.file)
		if err != nil {
			t.Errorf("could not decode %s: %v", test.file, err)
			continue
		}
		got := Convert(img)
		if got.W != test.want.W || got.H != test.want.H {
			t.Errorf("size does not match for %s got %dx%d want %dx%d", test.file,
				got.W, got.H, test.want.W, test.want.H,
			)
			continue
		}
		if !bytes.Equal(got.Data, test.want.Data) {
			t.Errorf("data does not match for %s\n got: %x\nwant: %x", test.file,
				got.Data, test.want.Data,
			)
		}
	}
}

func decImgFile(path string) (image.Image, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	return image.Decode(f)
}
