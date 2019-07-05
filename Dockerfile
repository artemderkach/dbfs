FROM golang:1.12.6-alpine3.9 AS builder

WORKDIR /srv/dbfs

COPY . .

RUN GO111MODULE=on go build -mod vendor

FROM alpine:3.9

COPY --from=builder /srv/dbfs/dbfs /srv/dbfs

ENTRYPOINT ["/srv/dbfs"]
