# trimmedmean

[![CI](https://github.com/dillonsingh26/trimmedmean/actions/workflows/ci.yml/badge.svg)](https://github.com/dillonsingh26/trimmedmean/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dillonsingh26/trimmedmean.svg)](https://pkg.go.dev/github.com/dillonsingh26/trimmedmean)

A small Go package for computing **symmetric and asymmetric trimmed means** of integer or floating-point slices.

A trimmed mean sorts the data, drops a proportion of the smallest and largest values, and averages what remains. It resists outliers better than the ordinary mean while using more of the data than the median.

Symmetric results match R's `mean(x, trim = p)` bit for bit, including R's rule of dropping `floor(n * p)` values from each end.

A companion repository, [trimmedmean-demo](https://github.com/dillonsingh26/trimmedmean-demo), shows the package in a complete program. It also compares results to R and runs a bootstrap study.

## Installation

You need Go 1.24 or newer. From inside your own Go module:

```bash
go get github.com/dillonsingh26/trimmedmean@latest
```

This downloads the package and records it in your `go.mod` file. If you don't have a module yet, create one first with `go mod init <your-module-name>`.

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/dillonsingh26/trimmedmean"
)

func main() {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 100}

	// One proportion: symmetric trimming, 10% from each end.
	// Same as mean(x, trim = 0.1) in R.
	symmetric, err := trimmedmean.TrimmedMean(data, 0.1)
	if err != nil {
		log.Fatal(err)
	}

	// Two proportions: asymmetric trimming, 0% from the low end and 10% from the high end.
	asymmetric, err := trimmedmean.TrimmedMean(data, 0, 0.1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(symmetric, asymmetric) // 6 5.5
}
```

### Function

```go
func TrimmedMean[T Number](data []T, trims ...float64) (float64, error)
```

| Call | Meaning |
|---|---|
| `TrimmedMean(data)` | No trimming (the ordinary mean) |
| `TrimmedMean(data, p)` | Symmetric: drop `floor(n*p)` values from each end |
| `TrimmedMean(data, low, high)` | Asymmetric: drop `floor(n*low)` from the bottom and `floor(n*high)` from the top |

- **`data`** can be a slice of any built-in integer or float type (`[]int`, `[]int64`, `[]uint8`, `[]float32`, `[]float64`, ...). Go generics make one function cover all of them.
- **The caller's slice is not changed.** The function sorts a copy.
- **The result is always `float64`.** Integers are converted to `float64` before averaging.

### Errors

The function returns an error instead of a misleading number. Check for each error with `errors.Is`:

| Error | Cause |
|---|---|
| `ErrEmptyData` | The data slice is empty |
| `ErrNaN` | The data contains `NaN` |
| `ErrTooManyProportions` | More than two proportions were given |
| `ErrInvalidProportion` | A proportion is negative or `NaN`, or the low and high proportions add up to 1 or more, which would leave nothing to average |

**Difference from R:** in R, `mean(x, trim = 0.5)` returns the median. Here, a symmetric proportion of 0.5 returns `ErrInvalidProportion`, because trimming half from each end leaves no values. Use a median function for that case.

## Testing

| Command | What it does |
|---|---|
| `make test` | Runs the unit tests, the fuzz seed inputs, and the documentation examples |
| `make fuzz` | Runs the fuzzer for 30 seconds |
| `make bench` | Measures speed and memory per call at 100, 10,000, and 1,000,000 values |
| `make cover` | Opens a coverage report in the browser (currently 100% of statements) |
| `make profile` | Profiles the benchmark and lists the functions using the most CPU |
| `make race` | Runs the tests with Go's data race detector |

Without `make`, you can run the same tests with `go test -v ./...`.

**The test files:**

- **`trimmedmean_test.go`**: unit tests.
  - Symmetric results are checked for exact equality against values printed by R's `mean(x, trim = p)`.
  - Asymmetric results are checked against R's `mean(sort(x)[(floor(n*lo)+1):(n-floor(n*hi))])`.
  - Also covers: every number type, invalid input, infinities, and that the input slice is left unchanged.
- **`fuzz_test.go`**: a fuzz test.
  - Go generates random data and proportions.
  - For each case, it checks that:
    - the function never crashes,
    - it returns an error exactly when the input is invalid,
    - its result exactly equals a simple reference version that adds the numbers as exact integers.
  - A local 30-second run tried about 6 million inputs with no failures.
- **`benchmark_test.go`**: the speed and memory benchmarks.
- **`example_test.go`**: runnable examples. Go checks their printed output, and they appear on the [package documentation page](https://pkg.go.dev/github.com/dillonsingh26/trimmedmean).

### Continuous integration

On every push, [GitHub Actions](.github/workflows/ci.yml) automatically:

- checks formatting,
- runs `go vet`,
- runs the tests with the race detector,
- fuzzes for 30 seconds,
- runs the benchmarks once.

The badge at the top of this page shows whether the latest commit passed.

### Performance

Measured on an Apple M4 with `make bench`:

| Values | Time per call | Memory per call | Allocations |
|---|---|---|---|
| 100 | 0.76 µs | 896 B | 1 |
| 10,000 | 0.21 ms | 80 KB | 1 |
| 1,000,000 | 71 ms | 8 MB | 1 |

The single allocation is the sorted copy of the data. Profiling (`make profile`) shows that sorting takes most of the time, while the averaging step takes about 2%.

## How it works

`TrimmedMean` works in four steps:

1. **Check the proportions.** No proportion means no trimming, and one proportion means symmetric trimming.
2. **Copy and sort the data.** The data is copied into a `float64` slice and sorted. Any `NaN` is rejected.
3. **Trim.** Drop `floor(n*low)` values from the bottom and `floor(n*high)` from the top.
4. **Average what's left, with rounding error kept to a minimum.**
   - Each `float64` addition rounds off a little, and the loss is largest when big outliers sit next to small values.
   - **Neumaier summation** keeps a second number that collects what each addition rounded off.
   - **`math.FMA`** recovers exactly what the final division rounded off, and that remainder is added back.

**Why step 4 matters.** A first version copied R's two-pass mean. R runs that method with 80-bit `long double` sums (on the x86_64 build used here, `.Machine$sizeof.longdouble` is 16 bytes), but Go's `float64` is 64-bit only. That first version was off from the exact answer by up to 45 units in the last digit (about 3e-13) on data with outliers.

The current method was checked against exact fraction arithmetic (Go's `math/big`). It gave the correctly rounded answer in every comparison sample and in 20,000 random data sets, and it matches R's printed results bit for bit.

## Files

| File | Role |
|---|---|
| `trimmedmean.go` | The package: `TrimmedMean`, its error values, and its helper functions |
| `trimmedmean_test.go` | Unit tests, checked against R |
| `fuzz_test.go` | Fuzz test |
| `benchmark_test.go` | Speed and memory benchmarks |
| `example_test.go` | Runnable documentation examples |
| `Makefile` | Shortcuts for testing, fuzzing, benchmarking, and profiling |
| `.github/workflows/ci.yml` | Automatic checks on GitHub |

## GenAI Tools

This package was built with help from Claude Code (Anthropic), an AI coding assistant running in the terminal. I used it to:

- summarize the assignment's packaging tutorial and point out where it is out of date,
- plan the package design,
- write the code, tests, and documentation, which I reviewed.

I also had it compute reference values in R so the tests check against R's own output. All tests, fuzzing, and benchmarks were run locally to confirm the results.

## License

[MIT](LICENSE)
