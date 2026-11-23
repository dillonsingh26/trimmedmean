# Common tasks. Run "make" with no target to format, vet, and test.

.PHONY: all fmt vet test race fuzz bench cover profile clean

all: fmt vet test

fmt:            ## reformat code to the standard Go style
	gofmt -s -w .

vet:            ## catch suspicious code the compiler allows
	go vet ./...

test:           ## run unit tests, the fuzz seeds, and the examples
	go test -v ./...

race:           ## run the tests with the data race detector
	go test -race ./...

fuzz:           ## try random inputs for 30 seconds, looking for failures
	go test -run '^$$' -fuzz FuzzTrimmedMean -fuzztime 30s

bench:          ## speed and memory per call at 100, 10,000, and 1,000,000 values
	go test -run '^$$' -bench . -benchmem

cover:          ## test coverage report, opened in the browser
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

profile:        ## CPU and memory profiles from the benchmark, then the top CPU users
	go test -run '^$$' -bench . -cpuprofile cpu.prof -memprofile mem.prof
	go tool pprof -top -nodecount=10 cpu.prof

clean:
	rm -f coverage.out *.prof *.test
