FROM golang:1.14
LABEL MAINTAINER="a nice guy"
WORKDIR $GOPATH/src/github.com/wybiral/tube
ADD . $GOPATH/src/github.com/wybiral/tube
RUN go env -w GOPROXY=https://goproxy.io && go get github.com/wybiral/feeds && go build -o tube  .
ENTRYPOINT ["./tube"]