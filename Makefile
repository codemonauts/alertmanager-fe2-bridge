build:
	CC=musl-gcc GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -linkmode external -extldflags -static"
