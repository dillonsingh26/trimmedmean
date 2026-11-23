package trimmedmean

import (
	"encoding/binary"
	"math"
	"slices"
	"testing"
)

// FuzzTrimmedMean feeds random data and proportions to TrimmedMean (run
// with make fuzz) and checks that:
//   - it never panics
//   - it returns an error exactly when the input is invalid
//   - the result exactly matches a slow reference computed with integer math
//   - one proportion gives the same result as two equal proportions
func FuzzTrimmedMean(f *testing.F) {
	// Seed inputs: Go mutates these to build new test cases.
	f.Add([]byte{1, 0, 2, 0, 3, 0, 100, 0}, 0.1, 0.1)
	f.Add([]byte{}, 0.05, 0.05)
	f.Add([]byte{255, 255, 0, 128, 7, 0}, 0.0, 0.4)
	f.Add([]byte{9, 0}, 0.5, 0.5)

	f.Fuzz(func(t *testing.T, raw []byte, lower, upper float64) {
		data := int16sFromBytes(raw)
		got, err := TrimmedMean(data, lower, upper)

		valid := len(data) > 0 && lower >= 0 && upper >= 0 && lower+upper < 1
		if !valid {
			if err == nil {
				t.Fatalf("expected an error for %d values, lower=%g, upper=%g", len(data), lower, upper)
			}
			return
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := referenceTrimmedMean(data, lower, upper)
		if got != want {
			t.Fatalf("got %.17g, reference %.17g (data=%v, lower=%g, upper=%g)", got, want, data, lower, upper)
		}

		if lower == upper {
			symmetric, _ := TrimmedMean(data, lower)
			if symmetric != got {
				t.Fatalf("one proportion gave %g, two equal proportions gave %g", symmetric, got)
			}
		}
	})
}

// int16sFromBytes reads pairs of bytes as signed 16-bit integers, giving
// the fuzzer a way to build slices with negative and repeated values.
func int16sFromBytes(raw []byte) []int16 {
	values := make([]int16, len(raw)/2)
	for i := range values {
		values[i] = int16(binary.LittleEndian.Uint16(raw[2*i:]))
	}
	return values
}

// referenceTrimmedMean is a deliberately simple version used to check the
// real one: sort, drop floor(n*p) from each end, and add with exact
// integer arithmetic. The only rounding is the final division, so the
// reference is the correctly rounded answer and must match bit for bit.
func referenceTrimmedMean(data []int16, lower, upper float64) float64 {
	sorted := slices.Clone(data)
	slices.Sort(sorted)

	n := len(sorted)
	kept := sorted[int(math.Floor(float64(n)*lower)) : n-int(math.Floor(float64(n)*upper))]

	var sum int64
	for _, v := range kept {
		sum += int64(v)
	}
	return float64(sum) / float64(len(kept))
}
