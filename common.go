package hyperloglog

import (
	"math/bits"

	"math"
)

type Hash32 interface {
	Sum32() uint32
}

type Hash64 interface {
	Sum64() uint64
}

type sortableSlice []uint32

func (p sortableSlice) Len() int           { return len(p) }
func (p sortableSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p sortableSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type set map[uint32]bool

func (s set) Add(i uint32) { s[i] = true }

func alpha(m uint32) float64 {
	if m == 16 {
		return 0.673
	} else if m == 32 {
		return 0.697
	} else if m == 64 {
		return 0.709
	}
	return 0.7213 / (1 + 1.079/float64(m))
}

func clz32(x uint32) uint8 {
	return uint8(bits.LeadingZeros32(x))
}

func clz64(x uint64) uint8 {
	return uint8(bits.LeadingZeros64(x))
}

// Extract bits from uint32 using LSB 0 numbering, including lo.
func eb32(bits uint32, hi uint8, lo uint8) uint32 {
	m := uint32(((1 << (hi - lo)) - 1) << lo)
	return (bits & m) >> lo
}

// Extract bits from uint64 using LSB 0 numbering, including lo.
func eb64(bits uint64, hi uint8, lo uint8) uint64 {
	m := uint64(((1 << (hi - lo)) - 1) << lo)
	return (bits & m) >> lo
}

func linearCounting(m uint32, v uint32) float64 {
	fm := float64(m)
	return fm * math.Log(fm/float64(v))
}

func countZeros(s []uint8) uint32 {
	var c uint32
	for _, v := range s {
		if v == 0 {
			c++
		}
	}
	return c
}

func calculateEstimate(s []uint8) float64 {
	if len(s) == 0 {
		return 0
	}

	sum := 0.0
	for _, val := range s {
		sum += 1.0 / float64(uint64(1)<<val)
	}

	m := uint32(len(s))
	fm := float64(m)
	return alpha(m) * fm * fm / sum
}
