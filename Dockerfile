FROM golang:1.19 as builder

ARG VERSION

WORKDIR /usr/app
COPY . ./

ENV CGO_ENABLED=0
RUN go build -o bin/service -ldflags="-X main.Version=${VERSION}" ./cmd

FROM alpine
WORKDIR /usr/app
COPY --from=builder /usr/app/bin/service ./service

ENTRYPOINT ["./service"]