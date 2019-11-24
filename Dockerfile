FROM golang:buster as builder

RUN apt update && apt install -y libmagickwand-6.q16-dev
RUN mkdir /build

WORKDIR /build

COPY *.go go.mod go.sum ./
COPY pkg ./pkg

RUN go build -o main

FROM debian:buster-slim

RUN apt update && apt install -y libmagickwand-6.q16-6 git

COPY --from=builder /build/main /

ENTRYPOINT ["/main"]
