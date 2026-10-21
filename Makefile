dev:
	air .

build:
	go build

run:
	go run .

test:
	go test -v ./...

prod:
	ngrok http --domain=sharp-seal-precise.ngrok-free.app 8069

cgo-build:
	env CGO_ENABLED=1 go build -tags enablecgo

cgo-run:
	env CGO_ENABLED=1 go run . -tags enablecgo
test
