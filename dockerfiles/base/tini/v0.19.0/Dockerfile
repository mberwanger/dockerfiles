# GENERATED FILE, DO NOT MODIFY!
# To update this file please edit the relevant source file and run the generation
# task 'go run cmd/main.go -generate tini'
FROM alpine

ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETVARIANT

RUN apk add --no-cache gnupg wget

# Install Tini for init use (reaps defunct processes and forwards signals)
RUN set -eux; \
  case "${TARGETARCH}" in \
    arm64|aarch64) export PLATFORM='-arm64' ;; \
    amd64|x86_64) export PLATFORM='' ;; \
    *) echo "Unsupported architecture: ${TARGETARCH}"; exit 1 ;; \
  esac; \
  wget -O /tini https://github.com/krallin/tini/releases/download/v0.19.0/tini${PLATFORM} ;\
  wget -O /tini.asc https://github.com/krallin/tini/releases/download/v0.19.0/tini${PLATFORM}.asc

ARG KEYSERVERS="keyserver.ubuntu.com keys.openpgp.org hkps://keyserver.ubuntu.com:443"
RUN --mount=type=cache,target=/var/cache/apk \
  set -eux; \
  for ks in ${KEYSERVERS}; do \
    echo "Trying keyserver: $ks"; \
    gpg --no-tty --keyserver-options timeout=10 --keyserver "${ks}" --recv-keys 595E85A6B1B4779EA4DAAEC70B588DFF0527A9B7 && break; \
  done \
  && gpg --no-tty --batch --verify /tini.asc /tini \
  && chmod +x /tini
