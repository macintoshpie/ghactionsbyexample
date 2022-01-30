
.PHONY: build
build:
	rm -rf public
	go get -t -v ./...
	go run generate.go
	cp -vr static public
