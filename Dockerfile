FROM golang:1.22.4-alpine3.20 AS builder
WORKDIR /go/src/int-mastodon
COPY . .
RUN \
    apk add protoc protobuf-dev make git && \
    make build

FROM alpine:3.20
RUN apk --no-cache add ca-certificates \
    && update-ca-certificates
COPY --from=builder /go/src/int-mastodon/int-mastodon /bin/int-mastodon
ENTRYPOINT ["/bin/int-mastodon"]
