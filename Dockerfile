FROM golang:1.14
LABEL MAINTAINER="a nice guy"
WORKDIR $GOPATH/src/github.com/wybiral/tube  #will be created if no exists
ADD static $GOPATH/src/github.com/wybiral/tube
ADD templates $GOPATH/src/github.com/wybiral/tube
ADD videos $GOPATH/src/github.com/wybiral/tube
ADD config.json $GOPATH/src/github.com/wybiral/tube
RUN go build -o tube  .

ENTRYPOINT ["./tube"]