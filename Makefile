build:
	go build -ldflags="-X github.com/snakehunterr/fsearch/args.VERSION=$(shell git describe --tags --abbrev=0)" -o fs .
