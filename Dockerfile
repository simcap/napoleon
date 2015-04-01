FROM golang

RUN mkdir -p /go/src/github.com/simcap/napoleon
ADD . /go/src/github.com/simcap/napoleon

RUN go install github.com/simcap/napoleon

WORKDIR /go/src/github.com/simcap/napoleon

ENTRYPOINT /go/bin/napoleon

EXPOSE 8080
