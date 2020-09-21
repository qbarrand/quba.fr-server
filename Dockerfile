FROM golang:alpine as builder

RUN apk add gcc imagemagick6-dev libc-dev
RUN mkdir /build

WORKDIR /build

COPY *.go go.mod go.sum ./
COPY pkg ./pkg

RUN go build -o main

FROM alpine

RUN apk add imagemagick6

COPY --from=builder /build/main /

ENTRYPOINT ["/main"]
