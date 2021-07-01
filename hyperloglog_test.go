package hyperloglog

import (
	"reflect"
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
	for p := 11; p <= 16; p++ {
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

func TestEB32(t *testing.T) {
	n := eb32(0xffffffff, 3, 1)
	if n != 3 {
		t.Error(n)
	}

	n = eb32(0xffffffff, 32, 0)
	if n != 0xffffffff {
		t.Error(n)
	}

	n = eb32(0xffffffff, 35, 0)
	if n != 0xffffffff {
		t.Error(n)
	}

	n = eb32(0xffffffff, 32, 10)
	if n != 0x3fffff {
		t.Error(n)
	}

	n = eb32(0xf001, 32, 16)
	if n != 0 {
		t.Error(n)
	}

	n = eb32(0xf001, 16, 0)
	if n != 0xf001 {
		t.Error(n)
	}

	n = eb32(0xf001, 12, 0)
	if n != 1 {
		t.Error(n)
	}

	n = eb32(0xf001, 16, 1)
	if n != 0x7800 {
		t.Error(n)
	}

	n = eb32(0x1211, 13, 2)
	if n != 0x484 {
		t.Error(n)
	}

	n = eb32(0x10000000, 32, 1)
	if n != 0x8000000 {
		t.Error(n)
	}
}
