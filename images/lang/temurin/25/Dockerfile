# GENERATED FILE, DO NOT MODIFY!
#
# To update this file please edit the relevant template file and run:
#   go run tool/main.go generate image temurin
#
# Or regenerate all images with:
#   go run tool/main.go generate all

ARG REGISTRY=ghcr.io/mberwanger
FROM ${REGISTRY}/core:noble

ARG JAVA_VERSION=25

USER root

RUN set -eux; \
    curl -fsSL --retry 3 --retry-delay 10 https://packages.adoptium.net/artifactory/api/gpg/key/public | gpg --dearmor | tee /etc/apt/trusted.gpg.d/adoptium.gpg > /dev/null \
 && echo "deb https://packages.adoptium.net/artifactory/deb $(awk -F= '/^VERSION_CODENAME/{print$2}' /etc/os-release) main" | tee /etc/apt/sources.list.d/adoptium.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends \
        temurin-${JAVA_VERSION}-jdk \
 && rm -rf /tmp/* /var/tmp/*

# Set JAVA_HOME and PATH
RUN ARCH=$(dpkg --print-architecture) && \
    ln -s /usr/lib/jvm/temurin-${JAVA_VERSION}-jdk-${ARCH} /usr/lib/jvm/temurin-${JAVA_VERSION}-jdk

ENV JAVA_HOME=/usr/lib/jvm/temurin-${JAVA_VERSION}-jdk
ENV PATH="/usr/lib/jvm/temurin-${JAVA_VERSION}-jdk/bin:${PATH}"

USER docker
