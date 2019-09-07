FROM golang:buster as builder

RUN apt update && apt install -y libmagickwand-6.q16-dev
RUN mkdir /build

WORKDIR /mnt

RUN go build -o /build/main

FROM debian:buster-slim

COPY --from=builder /build/main /

RUN apt update && apt install -y libmagickwand-6.q16-6

ENTRYPOINT ["/main"]
