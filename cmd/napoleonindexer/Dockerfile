FROM golang

RUN mkdir -p /go/src/github.com/simcap/napoleonindexer
ADD . /go/src/github.com/simcap/napoleonindexer
RUN go install github.com/simcap/napoleonindexer

RUN git clone https://github.com/steeve/france.code-civil.git

ENTRYPOINT ["napoleonindexer", "-r", "france.code-civil"]
