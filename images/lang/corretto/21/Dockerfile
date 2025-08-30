# GENERATED FILE, DO NOT MODIFY!
#
# To update this file please edit the relevant template file and run:
#   go run tool/main.go generate image corretto
#
# Or regenerate all images with:
#   go run tool/main.go generate all

ARG REGISTRY=ghcr.io/mberwanger
FROM ${REGISTRY}/core:noble

USER root

ARG JAVA_VERSION=21

# Install Corretto JDK in single layer with modern GPG handling
RUN set -eux; \
    mkdir -p /etc/apt/keyrings \
 && curl -fsSL --retry 3 https://apt.corretto.aws/corretto.key | gpg --dearmor > /etc/apt/keyrings/corretto.gpg \
 && echo "deb [signed-by=/etc/apt/keyrings/corretto.gpg] https://apt.corretto.aws stable main" > /etc/apt/sources.list.d/corretto.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends \
        java-${JAVA_VERSION}-amazon-corretto-jdk \
 && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Set JAVA_HOME and PATH
ENV JAVA_HOME=/usr/lib/jvm/java-${JAVA_VERSION}-amazon-corretto \
    PATH=/usr/lib/jvm/java-${JAVA_VERSION}-amazon-corretto/bin:${PATH}

USER docker
