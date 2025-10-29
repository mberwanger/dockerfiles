# GENERATED FILE, DO NOT MODIFY!
#
# To update this file please edit the relevant template file and run:
#   go run tool/main.go generate image tini
#
# Or regenerate all images with:
#   go run tool/main.go generate all

FROM alpine:3.22

RUN apk add --no-cache gnupg curl

# Install Tini for init use (reaps defunct processes and forwards signals)
RUN apkArch="$(apk --print-arch)"; \
  case "${apkArch}" in \
    aarch64) export PLATFORM='-arm64' ;; \
    x86_64) export PLATFORM='' ;; \
  esac; \
  curl -fsSL --retry 3 --retry-delay 10 -o /tini https://github.com/krallin/tini/releases/download/v0.19.0/tini${PLATFORM} ;\
  curl -fsSL --retry 3 --retry-delay 10 -o /tini.asc https://github.com/krallin/tini/releases/download/v0.19.0/tini${PLATFORM}.asc

ARG KEYSERVERS="keyserver.ubuntu.com keys.openpgp.org hkps://keyserver.ubuntu.com:443"
RUN --mount=type=cache,target=/var/cache/apk \
  set -eux; \
  for ks in ${KEYSERVERS}; do \
    echo "Trying keyserver: ${ks}"; \
    gpg --no-tty --keyserver-options timeout=10 --keyserver "${ks}" --recv-keys 595E85A6B1B4779EA4DAAEC70B588DFF0527A9B7 && break; \
  done \
  && gpg --no-tty --batch --verify /tini.asc /tini \
  && chmod +x /tini
