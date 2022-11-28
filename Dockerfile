FROM golang:1.19.1-alpine as go-builder
WORKDIR /repo-scanner
COPY internal internal
COPY cmd cmd
COPY go.mod .
COPY go.sum .
COPY Makefile .

ARG VERSION="latest"
ENV VERSION="$VERSION"

RUN echo "$VERSION"
RUN apk add build-base &&\
    make clean && make && \
    chmod 777 build/server

FROM alpine:latest
RUN adduser -D -h /repo-scanner -u 1000 -k /dev/null repo-scanner
WORKDIR /repo-scanner
COPY --from=go-builder --chown=nobody:nobody /repo-scanner/build build
COPY --chown=nobody:nobody .env .
RUN apk add git
EXPOSE 8080
USER repo-scanner
CMD [ "./build/server" ]

