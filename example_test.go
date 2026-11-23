package trimmedmean_test

import (
	"fmt"

	"github.com/dillonsingh26/trimmedmean"
)

// One proportion trims the same share from both ends, like R's
// mean(x, trim = 0.1).
func ExampleTrimmedMean_symmetric() {
	scores := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100}

	average, _ := trimmedmean.TrimmedMean(scores)
	trimmed, _ := trimmedmean.TrimmedMean(scores, 0.1)

	fmt.Printf("mean: %.2f\n", average)
	fmt.Printf("10%% trimmed mean: %.2f\n", trimmed)
	// Output:
	// mean: 14.09
	// 10% trimmed mean: 6.00
}

// Two proportions trim the low and high ends separately. Here only the
// high end is trimmed, removing the large outlier.
func ExampleTrimmedMean_asymmetric() {
	responseTimes := []float64{1.2, 0.8, 1.1, 0.9, 1.0, 1.3, 0.7, 1.0, 1.2, 0.9, 30.5}

	trimmed, err := trimmedmean.TrimmedMean(responseTimes, 0, 0.1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("high-end trimmed mean: %.2f\n", trimmed)
	// Output:
	// high-end trimmed mean: 1.01
}

// Invalid proportions return an error instead of a misleading number.
func ExampleTrimmedMean_invalid() {
	_, err := trimmedmean.TrimmedMean([]float64{1, 2, 3}, 0.5)
	fmt.Println(err)
	// Output:
	// trimmedmean: invalid trimming proportion: lower=0.5, upper=0.5 (each must be at least 0, and their sum below 1)
}
