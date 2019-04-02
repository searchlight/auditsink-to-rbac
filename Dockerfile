FROM golang

COPY . /go/src/github.com/masudur-rahman/audit-sink

RUN go install /go/src/github.com/masudur-rahman/audit-sink

ENTRYPOINT ["/go/bin/audit-sink"]

EXPOSE 4000
