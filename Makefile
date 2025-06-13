build:
	go build -o fs .
install:
	go build -o fs . && sudo mv ./fs /usr/bin/fs
