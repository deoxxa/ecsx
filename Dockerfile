FROM golang:1.6

RUN mkdir -p /go/src/fknsrs.biz/p/ecsx

WORKDIR /go/src/fknsrs.biz/p/ecsx

ADD vendor vendor/
ADD *.go ./

RUN go install

ENTRYPOINT ["/go/bin/ecsx"]
