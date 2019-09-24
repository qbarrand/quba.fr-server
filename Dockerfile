FROM golang:buster as builder

RUN apt update && apt install -y libmagickwand-6.q16-dev git
RUN mkdir /build

WORKDIR /build

COPY *.go go.mod go.sum ./

RUN go build -o main

FROM debian:buster-slim

RUN apt update && apt install -y libmagickwand-6.q16-6

COPY --from=builder /build/main /

ENTRYPOINT ["/main"]
