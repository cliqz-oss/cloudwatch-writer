FROM golang:1.10-alpine

# setup dep
RUN apk update && apk add git
RUN wget -O /bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod a+x /bin/dep

RUN mkdir -p /go/src/github.com/cliqz/cloudwatch-writer/
WORKDIR /go/src/github.com/cliqz/cloudwatch-writer/
ADD . ./
RUN dep ensure && go build && go install

ENTRYPOINT ["/go/bin/cloudwatch-writer"]

