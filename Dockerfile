FROM golang:1.17.8-alpine3.14 AS build-env
RUN apk add --no-cache build-base
RUN go install github.com/jaeles-project/gospider@latest

FROM alpine:3.15.0
RUN apk add --no-cache bind-tools ca-certificates
COPY --from=build-env /go/bin/gospider /usr/local/bin/gospider
ENTRYPOINT ["gospider"]
