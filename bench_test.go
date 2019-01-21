package hyperloglog

import (
	"fmt"
	"hash"
	"hash/fnv"
	"math"
	"math/rand"
	"testing"
)

func hash32(s string) hash.Hash32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h
}

func hash64(s string) hash.Hash64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h
}

func randStr(n int) string {
	i := rand.Uint32()
	return fmt.Sprintf("%d %d", i, n)
}

func benchmark(b *testing.B, precision uint8) {
	h, _ := New(precision)
	hpp, _ := NewPlus(precision)

	items := make([]string, b.N)
	for i := 0; i < len(items); i++ {
		items[i] = randStr(i)
	}

	b.ResetTimer()
	for _, s := range items {
		h.Add(hash32(s))
		h.Add(hash32(s))
		hpp.Add(hash64(s))
		hpp.Add(hash64(s))
	}
	b.StopTimer()

	e, epp := h.Count(), hpp.Count()

	var percentErr = func(est uint64) float64 {
		return math.Abs(float64(b.N)-float64(est)) / float64(b.N)
	}

	fmt.Printf("\nReal Cardinality: %8d\n", b.N)
	fmt.Printf("HyperLogLog     : %8d,   Error: %f%%\n", e, percentErr(e))
	fmt.Printf("HyperLogLog++   : %8d,   Error: %f%%\n", epp, percentErr(epp))
}

func BenchmarkHll4(b *testing.B) {
	benchmark(b, 4)
}

func BenchmarkHll6(b *testing.B) {
	benchmark(b, 6)
}

func BenchmarkHll8(b *testing.B) {
	benchmark(b, 8)
}

func BenchmarkHll10(b *testing.B) {
	benchmark(b, 10)
}

func BenchmarkHll14(b *testing.B) {
	benchmark(b, 14)
}

func BenchmarkHll16(b *testing.B) {
	benchmark(b, 16)
}
