FROM golang:latest AS build-env
RUN GO111MODULE=on go install github.com/jaeles-project/gospider@latest
FROM alpine:3.17.1
RUN apk add --no-cache ca-certificates libc6-compat
WORKDIR /app
COPY --from=build-env /go/bin/gospider .
RUN mkdir -p /app \
    && adduser -D gospider \
    && chown -R gospider:gospider /app
USER gospider
ENTRYPOINT [ "./gospider" ]