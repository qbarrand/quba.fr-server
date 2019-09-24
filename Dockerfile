FROM golang:buster as builder

RUN apt update && apt install -y libmagickwand-6.q16-dev git
RUN mkdir /build

COPY main.go server.go go.mod go.sum /build/

WORKDIR /build

RUN go build -o main

FROM debian:buster-slim

COPY --from=builder /build/main /

RUN apt update && apt install -y libmagickwand-6.q16-6

ENTRYPOINT ["/main"]
