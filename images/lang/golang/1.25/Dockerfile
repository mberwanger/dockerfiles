# GENERATED FILE, DO NOT MODIFY!
#
# To update this file please edit the relevant template file and run:
#   go run tool/main.go generate image golang
#
# Or regenerate all images with:
#   go run tool/main.go generate all

ARG REGISTRY=ghcr.io/mberwanger
FROM ${REGISTRY}/core:noble

USER root

ENV GOLANG_VERSION=1.25.1

# Multi-architecture support
RUN set -eux; \
    dpkgArch="$(dpkg --print-architecture)"; \
    case "${dpkgArch}" in \
        amd64) \
            GOLANG_DOWNLOAD_URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"; \
            GOLANG_DOWNLOAD_SHA256="7716a0d940a0f6ae8e1f3b3f4f36299dc53e31b16840dbd171254312c41ca12e"; \
            ;; \
        arm64) \
            GOLANG_DOWNLOAD_URL="https://go.dev/dl/go${GOLANG_VERSION}.linux-arm64.tar.gz"; \
            GOLANG_DOWNLOAD_SHA256="65a3e34fb2126f55b34e1edfc709121660e1be2dee6bdf405fc399a63a95a87d"; \
            ;; \
        *) echo >&2 "unsupported architecture: ${dpkgArch}"; exit 1 ;; \
    esac; \
    curl -fsSL --retry 3 "${GOLANG_DOWNLOAD_URL}" -o golang.tar.gz \
 && echo "${GOLANG_DOWNLOAD_SHA256}  golang.tar.gz" | sha256sum -c - \
 && tar -C /usr/local -xzf golang.tar.gz \
 && rm golang.tar.gz \
 && /usr/local/go/bin/go version

# Install git and make for Go development
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
        git \
        make \
 && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENV GOPATH=/go
ENV PATH=${GOPATH}/bin:/usr/local/go/bin:${PATH}
ENV SRCPATH=${GOPATH}/src
ENV GOPRIVATE github.com/mberwanger

RUN mkdir -p ${SRCPATH} && chown -R docker:docker ${GOPATH}

WORKDIR ${SRCPATH}
USER docker
