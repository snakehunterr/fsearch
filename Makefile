build:
	go build -ldflags="-X main.version=$(shell git describe --tags --abbrev=0)" -o fs .
