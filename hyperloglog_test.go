package hyperloglog

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/spaolacci/murmur3"
)

type fakeHash32 uint32

func (f fakeHash32) Sum32() uint32 { return uint32(f) }

func TestHLLAdd(t *testing.T) {
	h, _ := New(16)

	h.Add(fakeHash32(0x00010fff))
	n := h.reg[1]
	if n != 5 {
		t.Error(n)
	}

	h.Add(fakeHash32(0x0002ffff))
	n = h.reg[2]
	if n != 1 {
		t.Error(n)
	}

	h.Add(fakeHash32(0x00030000))
	n = h.reg[3]
	if n != 17 {
		t.Error(n)
	}

	h.Add(fakeHash32(0x00030001))
	n = h.reg[3]
	if n != 17 {
		t.Error(n)
	}

	h.Add(fakeHash32(0xff037000))
	n = h.reg[0xff03]
	if n != 2 {
		t.Error(n)
	}

	h.Add(fakeHash32(0xff030800))
	n = h.reg[0xff03]
	if n != 5 {
		t.Error(n)
	}
}

func TestHLLCount(t *testing.T) {
	var empty HyperLogLog
	if c := empty.Count(); c != 0 {
		t.Errorf("zero-initiated hll count: %d", c)
	}

	// TODO: make this test pass for smaller p
	for p := byte(11); p <= 16; p++ {
		h, _ := New(p)

		n := h.Count()
		if n != 0 {
			t.Error(n)
		}

		for i := 1; i <= 8; i++ {
			var b [16]byte
			b[0] = byte(i)
			h.Add(fakeHash32(murmur3.Sum32(b[:])))

			n = h.Count()
			if n != uint64(i) {
				t.Error(p, n, i)
			}
		}
	}
}

func TestHLLMergeError(t *testing.T) {
	h, _ := New(16)
	h2, _ := New(10)

	err := h.Merge(h2)
	if err == nil {
		t.Error("different precision should return error")
	}
}

func TestHLLMerge(t *testing.T) {
	h, _ := New(16)
	h.Add(fakeHash32(0x00010fff))
	h.Add(fakeHash32(0x00020fff))
	h.Add(fakeHash32(0x00030fff))
	h.Add(fakeHash32(0x00040fff))
	h.Add(fakeHash32(0x00050fff))
	h.Add(fakeHash32(0x00050fff))

	h2, _ := New(16)
	h2.Merge(h)
	n := h2.Count()
	if n != 5 {
		t.Error(n)
	}

	h2.Merge(h)
	n = h2.Count()
	if n != 5 {
		t.Error(n)
	}

	h.Add(fakeHash32(0x00060fff))
	h.Add(fakeHash32(0x00070fff))
	h.Add(fakeHash32(0x00080fff))
	h.Add(fakeHash32(0x00090fff))
	h.Add(fakeHash32(0x000a0fff))
	h.Add(fakeHash32(0x000a0fff))
	n = h.Count()
	if n != 10 {
		t.Error(n)
	}

	h2.Merge(h)
	n = h2.Count()
	if n != 10 {
		t.Error(n)
	}
}

func TestHLLClear(t *testing.T) {
	h, _ := New(16)
	h.Add(fakeHash32(0x00010fff))

	n := h.Count()
	if n != 1 {
		t.Error(n)
	}
	h.Clear()

	n = h.Count()
	if n != 0 {
		t.Error(n)
	}

	h.Add(fakeHash32(0x00010fff))
	n = h.Count()
	if n != 1 {
		t.Error(n)
	}
}

func TestHLLPrecision(t *testing.T) {
	h, _ := New(4)

	h.Add(fakeHash32(0x1fffffff))
	n := h.reg[1]
	if n != 1 {
		t.Error(n)
	}

	h.Add(fakeHash32(0xffffffff))
	n = h.reg[0xf]
	if n != 1 {
		t.Error(n)
	}

	h.Add(fakeHash32(0x00ffffff))
	n = h.reg[0]
	if n != 5 {
		t.Error(n)
	}
}

func TestHLLError(t *testing.T) {
	_, err := New(3)
	if err == nil {
		t.Error("precision 3 should return error")
	}

	_, err = New(17)
	if err == nil {
		t.Error("precision 17 should return error")
	}
}

func TestHLLGob(t *testing.T) {
	var c1, c2 struct {
		HLL   *HyperLogLog
		Count int
	}
	c1.HLL, _ = New(8)
	for _, h := range []fakeHash32{0x10fff, 0x20fff, 0x30fff, 0x40fff, 0x50fff} {
		c1.HLL.Add(h)
		c1.Count++
	}

	var buf bytes.Buffer

	if err := gob.NewEncoder(&buf).Encode(&c1); err != nil {
		t.Error(err)
	}
	if err := gob.NewDecoder(&buf).Decode(&c2); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(c1, c2) {
		t.Error("unmarshaled structure differs")
	}
}

func TestHLLNewReg(t *testing.T) {
	_, err := NewReg(make([]uint8, 1<<3))
	if err == nil {
		t.Error("expected error")
	}
	_, err = NewReg(make([]uint8, 1<<17))
	if err == nil {
		t.Error("expected error")
	}

	h, _ := New(16)
	h.Add(fakeHash32(0x00010fff))
	h.Add(fakeHash32(0x00020fff))
	h.Add(fakeHash32(0x00030fff))
	h.Add(fakeHash32(0x00040fff))
	h.Add(fakeHash32(0x00050fff))
	h.Add(fakeHash32(0x00050fff))

	h2, err := NewReg(h.Registers())
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(h, h2) {
		t.Error("HLLs differs")
	}
}

func TestHLL_TextMarshaler(t *testing.T) {
	h, _ := New(10)
	h.Add(fakeHash32(0x00010fff))
	h.Add(fakeHash32(0x00020fff))
	h.Add(fakeHash32(0x00030fff))
	h.Add(fakeHash32(0x00040fff))
	h.Add(fakeHash32(0x00050fff))
	h.Add(fakeHash32(0x00050fff))

	txt, err := h.MarshalText()
	t.Logf("hll as text: %s", txt)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("txt size:", len(txt))

	h2 := new(HyperLogLog)
	if err := h2.UnmarshalText(txt); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(h, h2) {
		t.Fatalf("HLLs differs:\n%v\n%v", h.p, h2.p)
	}
}

// make sure JSON uses text marshaler
func TestHLL_JSON(t *testing.T) {
	h, _ := New(10)
	h.Add(fakeHash32(0x00010fff))
	h.Add(fakeHash32(0x00020fff))
	h.Add(fakeHash32(0x00030fff))
	h.Add(fakeHash32(0x00040fff))
	h.Add(fakeHash32(0x00050fff))
	h.Add(fakeHash32(0x00050fff))

	// Get text representation of HLL
	text, err := h.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	// Encode JSON
	jd, err := json.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Compare(fmt.Sprintf(`"%s"`, text), string(jd)) != 0 {
		t.Fatalf("%s", jd)
	}
}
