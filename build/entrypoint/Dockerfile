FROM golang:1.19.5-alpine3.17 as base
FROM base as built
WORKDIR /go/app/api
COPY ../../go.mod .
COPY ../../go.sum .
COPY ../../cmd/entrypoint .
ENV CGO_ENABLED=0

RUN go get -d -v ./...
RUN go build -o /tmp/entrypoint

FROM alpine:3.17 AS certificates
RUN apk --no-cache add ca-certificates

FROM busybox
COPY --from=built /tmp/entrypoint /usr/bin/entrypoint
COPY --from=certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 8000

CMD ["entrypoint", "start"]
