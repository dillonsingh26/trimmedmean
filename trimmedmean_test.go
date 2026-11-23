package trimmedmean

import (
	"errors"
	"math"
	"slices"
	"testing"
)

// TestMatchesR checks symmetric trimming against values printed by R's
// mean(x, trim = p), and asymmetric trimming against
// mean(sort(x)[(floor(n*lo)+1):(n-floor(n*hi))]). R printed 17 significant
// digits, so the comparison is exact: every bit must match.
func TestMatchesR(t *testing.T) {
	tests := []struct {
		name  string
		data  []float64
		trims []float64
		want  float64
	}{
		{"outlier trimmed", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100}, []float64{0.1}, 6},
		{"no trim is the ordinary mean", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100}, nil, 14.090909090909092},
		{"floor(10 * 0.05) = 0 drops nothing", []float64{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}, []float64{0.05}, 3.9},
		{"floor(7 * 0.25) = 1 from each end", []float64{-7, 12, 0, 3, -2, 8, 40}, []float64{0.25}, 4.2},
		{"fractional values", []float64{1.5, 2.25, -3.75, 1e6, 0.125}, []float64{0.2}, 1.2916666666666667},
		{"decimals that are inexact in binary", []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0}, []float64{0.1}, 0.55},
		{"extreme values at both ends", []float64{-1e6, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 1e6}, []float64{0.05}, 9.5},
		{"single value", []float64{5}, []float64{0.4}, 5},
		{"asymmetric: high end only", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100}, []float64{0, 0.1}, 5.5},
		{"asymmetric: 2 low, 5 high of 20", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, []float64{0.1, 0.25}, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TrimmedMean(tt.data, tt.trims...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %.17g, want %.17g", got, tt.want)
			}
		})
	}
}

// TestOneProportionIsSymmetric checks that passing p once gives the same
// result as passing it for both ends.
func TestOneProportionIsSymmetric(t *testing.T) {
	data := []int{12, -4, 7, 7, 30, 1, 0, 9, 15, 3, 88, -20}
	for _, p := range []float64{0, 0.05, 0.1, 0.25, 0.49} {
		once, err1 := TrimmedMean(data, p)
		twice, err2 := TrimmedMean(data, p, p)
		if err1 != nil || err2 != nil {
			t.Fatalf("p=%g: unexpected errors %v, %v", p, err1, err2)
		}
		if once != twice {
			t.Errorf("p=%g: one proportion gave %g, two equal proportions gave %g", p, once, twice)
		}
	}
}

// TestNumberTypes checks that every kind of number gives the same answer.
func TestNumberTypes(t *testing.T) {
	const want = 6.0 // R: mean(c(1:10, 100), trim = 0.1)
	results := map[string]func() (float64, error){
		"int":     func() (float64, error) { return TrimmedMean([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100}, 0.1) },
		"int32":   func() (float64, error) { return TrimmedMean([]int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100}, 0.1) },
		"uint8":   func() (float64, error) { return TrimmedMean([]uint8{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100}, 0.1) },
		"float32": func() (float64, error) { return TrimmedMean([]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100}, 0.1) },
	}
	for name, compute := range results {
		got, err := compute()
		if err != nil || got != want {
			t.Errorf("%s: got %g, %v; want %g", name, got, err, want)
		}
	}
}

func TestDoesNotModifyInput(t *testing.T) {
	data := []float64{9, 1, 8, 2, 7}
	original := slices.Clone(data)
	if _, err := TrimmedMean(data, 0.2); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(data, original) {
		t.Errorf("input changed from %v to %v", original, data)
	}
}

// TestInfinityTrimmedAway shows trimming removes values that would
// otherwise make the ordinary mean infinite.
func TestInfinityTrimmedAway(t *testing.T) {
	data := []float64{math.Inf(-1), 1, 2, 3, 4, 5, 6, 7, 8, math.Inf(1)}
	got, err := TrimmedMean(data, 0.1)
	if err != nil || got != 4.5 {
		t.Errorf("got %g, %v; want 4.5", got, err)
	}
}

// TestInfinityKept matches R, where mean(c(1, 2, Inf)) is Inf, not NaN.
func TestInfinityKept(t *testing.T) {
	got, err := TrimmedMean([]float64{1, 2, math.Inf(1)})
	if err != nil || !math.IsInf(got, 1) {
		t.Errorf("got %g, %v; want +Inf", got, err)
	}
}

func TestInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		data    []float64
		trims   []float64
		wantErr error
	}{
		{"empty data", nil, []float64{0.1}, ErrEmptyData},
		{"NaN in data", []float64{1, math.NaN(), 3}, []float64{0.1}, ErrNaN},
		{"three proportions", []float64{1, 2, 3}, []float64{0.1, 0.1, 0.1}, ErrTooManyProportions},
		{"negative proportion", []float64{1, 2, 3}, []float64{-0.1}, ErrInvalidProportion},
		{"symmetric 0.5 leaves nothing", []float64{1, 2, 3}, []float64{0.5}, ErrInvalidProportion},
		{"asymmetric sum of 1", []float64{1, 2, 3}, []float64{0.3, 0.7}, ErrInvalidProportion},
		{"NaN proportion", []float64{1, 2, 3}, []float64{math.NaN()}, ErrInvalidProportion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := TrimmedMean(tt.data, tt.trims...)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}
