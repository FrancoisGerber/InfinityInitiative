dev:
	air .

build:
	go build

test:
	go test -v ./...

# swagger:
# 	swag init -g ./app.go -o ./api/docs
#
# deploy: swagger build test
#
# run: swagger build dev
