
.PHONY: build
build:
	rm -rf public
	go get -t -v ./...
	go run generate.go
	rsync -rupE static public
