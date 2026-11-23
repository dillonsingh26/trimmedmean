package trimmedmean

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

// BenchmarkTrimmedMean measures time and memory per call as the data grows (run with make bench).
func BenchmarkTrimmedMean(b *testing.B) {
	for _, size := range []int{100, 10_000, 1_000_000} {
		data := randomFloats(size)
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := TrimmedMean(data, 0.05); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// randomFloats returns normally distributed values from a fixed seed, so
// every benchmark run measures the same data.
func randomFloats(size int) []float64 {
	rng := rand.New(rand.NewPCG(1, 2))
	values := make([]float64, size)
	for i := range values {
		values[i] = 100 + 10*rng.NormFloat64()
	}
	return values
}
