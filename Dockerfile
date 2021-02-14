FROM golang:1.14

# setup
RUN apt-get update && apt-get install -y python3 python3-pip
RUN python3 -m pip install Pygments

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...

CMD ["go", "run", "generate.go"]
