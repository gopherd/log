all: genslice loglint benchmark

genslice:
	go run genslice.go > slice.go

loglint:
	cd cmd/loglint && go install

test:
	go test

benchmark:
	go test -bench=. -benchmem
