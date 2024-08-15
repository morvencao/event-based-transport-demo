FROM golang:1.22-bullseye AS builder
ARG OS=linux
ARG ARCH=amd64
WORKDIR /go/src/github.com/morvencao/event-based-transport-demo
COPY . .
ENV GO_PACKAGE github.com/morvencao/event-based-transport-demo

RUN GOOS=${OS} \
    GOARCH=${ARCH} \
    make build --warn-undefined-variables

FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
ENV USER_UID=10001

COPY --from=builder /go/src/github.com/morvencao/event-based-transport-demo/event-based-transport-demo /

USER ${USER_UID}
