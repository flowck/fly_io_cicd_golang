FROM golang:1.19 as builder

ARG VERSION

WORKDIR /usr/app
COPY . ./

ENV CGO_ENABLED=0
RUN go build -buildvcs=false -o bin/service -ldflags="-X main.Version=${VERSION}" ./cmd

FROM alpine
WORKDIR /usr/app
RUN mkdir misc misc/sql misc/sql/migrations
COPY --from=builder /usr/app/misc/sql/migrations/* ./misc/sql/migrations/
COPY --from=builder /usr/app/bin/service ./service

ENTRYPOINT ["./service"]