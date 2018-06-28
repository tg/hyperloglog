// Package hyperloglog implements the HyperLogLog and HyperLogLog++ cardinality
// estimation algorithms.
// These algorithms are used for accurately estimating the cardinality of a
// multiset using constant memory. HyperLogLog++ has multiple improvements over
// HyperLogLog, with a much lower error rate for smaller cardinalities.
//
// HyperLogLog is described here:
// http://algo.inria.fr/flajolet/Publications/FlFuGaMe07.pdf
//
// HyperLogLog++ is described here:
// http://research.google.com/pubs/pub40671.html
package hyperloglog

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"math"
)

const two32 = 1 << 32

type HyperLogLog struct {
	reg []uint8
	m   uint32
	p   uint8
}

// New returns a new initialized HyperLogLog.
func New(precision uint8) (*HyperLogLog, error) {
	if precision > 16 || precision < 4 {
		return nil, errors.New("precision must be between 4 and 16")
	}

	h := &HyperLogLog{}
	h.p = precision
	h.m = 1 << precision
	h.reg = make([]uint8, h.m)
	return h, nil
}

// NewReg creates HyperLogLog, which uses passed array as registers.
func NewReg(reg []uint8) (*HyperLogLog, error) {
	if len(reg) > (1<<16) || len(reg) < (1<<4) {
		return nil, errors.New("number of registers out of range")
	}

	p := 31 - clz32(uint32(len(reg)))
	m := 1 << uint32(p)
	if m != len(reg) {
		return nil, errors.New("invalid number of registers (must be power of 2)")
	}

	h := &HyperLogLog{
		p:   uint8(p),
		m:   uint32(m),
		reg: reg,
	}

	return h, nil
}

// Registers returns raw registers of HyperLogLog.
func (h *HyperLogLog) Registers() []uint8 {
	return h.reg
}

// Copy returns copy of HyperLogLog.
func (h *HyperLogLog) Copy() *HyperLogLog {
	reg := make([]byte, len(h.reg))
	copy(reg, h.reg)
	h, err := NewReg(reg)
	if err != nil {
		// should never happen assuming h is correct
		panic(err)
	}
	return h
}

// Clear sets HyperLogLog h back to its initial state.
func (h *HyperLogLog) Clear() {
	h.reg = make([]uint8, h.m)
}

// Add adds a new item to HyperLogLog h.
func (h *HyperLogLog) Add(item Hash32) {
	x := item.Sum32()
	i := eb32(x, 32, 32-h.p) // {x31,...,x32-p}
	w := x<<h.p | 1<<(h.p-1) // {x32-p,...,x0}

	zeroBits := clz32(w) + 1
	if zeroBits > h.reg[i] {
		h.reg[i] = zeroBits
	}
}

// Merge takes another HyperLogLog and combines it with HyperLogLog h.
func (h *HyperLogLog) Merge(other *HyperLogLog) error {
	if h.p != other.p {
		return errors.New("precisions must be equal")
	}

	for i, v := range other.reg {
		if v > h.reg[i] {
			h.reg[i] = v
		}
	}
	return nil
}

// Count returns the cardinality estimate.
func (h *HyperLogLog) Count() uint64 {
	est := calculateEstimate(h.reg)
	if est <= float64(h.m)*2.5 {
		if v := countZeros(h.reg); v != 0 {
			return uint64(linearCounting(h.m, v))
		}
		return uint64(est)
	} else if est < two32/30 {
		return uint64(est)
	}
	return uint64(-two32 * math.Log(1-est/two32))
}

// Encode HyperLogLog into a gob
func (h *HyperLogLog) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(h.reg); err != nil {
		return nil, err
	}
	if err := enc.Encode(h.m); err != nil {
		return nil, err
	}
	if err := enc.Encode(h.p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode gob into a HyperLogLog structure
func (h *HyperLogLog) GobDecode(b []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	if err := dec.Decode(&h.reg); err != nil {
		return err
	}
	if err := dec.Decode(&h.m); err != nil {
		return err
	}
	if err := dec.Decode(&h.p); err != nil {
		return err
	}
	return nil
}

// MarshalText marshals HLL into text data (registers as base64)
func (h *HyperLogLog) MarshalText() ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(h.reg)))
	base64.StdEncoding.Encode(dst, h.reg)
	return dst, nil
}

// UnmarshalText unmarshals HLL from text data prod by MarshalText
func (h *HyperLogLog) UnmarshalText(text []byte) error {
	reg := make([]byte, base64.StdEncoding.DecodedLen(len(text)))
	n, err := base64.StdEncoding.Decode(reg, text)
	if err != nil {
		return err
	}
	h2, err := NewReg(reg[:n])
	if err != nil {
		return err
	}
	*h = *h2
	return nil
}
