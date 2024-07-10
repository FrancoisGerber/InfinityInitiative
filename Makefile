dev:
	air .

build:
	go build

run:
	go run .

test:
	go test -v ./...

cgo-build:
	env CGO_ENABLED=1 go build -tags enablecgo

cgo-run:
	env CGO_ENABLED=1 go run . -tags enablecgo
