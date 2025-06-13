build:
	git describe --tags --abbrev=0 > args/version
	go build -o fs .
install:
	git describe --tags --abbrev=0 > args/version
	go build -o fs . && sudo mv ./fs /usr/bin/fs
