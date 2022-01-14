# syntax = docker/dockerfile:1.2

# == base ==

FROM --platform=$BUILDPLATFORM golang:1.16-buster AS base
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.* .
RUN --mount=type=cache,target=/go/pkg/mod go mod download

FROM scratch AS base-copy
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=latest

# == bin ==

FROM base AS build
ENV GOOS=${TARGETOS} GOARCH=${TARGETARCH}
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /dist/ossenc .

FROM base-copy AS dist-base-windows
ENV SUFFIX=.exe

FROM base-copy AS dist-base-linux

FROM dist-base-${TARGETOS} as dist
COPY --from=build /dist/ossenc /${VERSION}/ossenc-${VERSION}-${TARGETOS}-${TARGETARCH}${SUFFIX}
