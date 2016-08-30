FROM golang:1.7-alpine
ADD . /go/src/github.mit.edu/tep/lates
WORKDIR /go/src/github.mit.edu/tep/lates
EXPOSE 8080
ENTRYPOINT go run lates.go

