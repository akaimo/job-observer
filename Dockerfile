FROM golang:1.14.2 as builder

WORKDIR /go/src/github.com/akaimo.com/job-observer

COPY go.mod .
COPY go.sum .

RUN set -x \
  && go mod download

COPY . .

RUN set -x \
  && go build -o job-observer ./cmd/controller

FROM gcr.io/distroless/base-debian10

COPY --from=builder /go/src/github.com/akaimo.com/job-observer/job-observer /

CMD ["/job-observer"]
