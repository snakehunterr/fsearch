build:
	go build -ldflags="-X github.com/snakehunterr/fsearch/args.VERSION=$(shell git tag | sort -r | head -1)" -o fs .
