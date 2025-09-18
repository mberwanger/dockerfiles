# GENERATED FILE, DO NOT MODIFY!
# To update this file please edit the relevant source file and run the generation
# task 'go run cmd/main.go -generate corretto'
ARG REGISTRY=ghcr.io/mberwanger
FROM ${REGISTRY}/core:bullseye

LABEL org.opencontainers.image.authors="Martin Berwanger" \
      org.opencontainers.image.title="Amazon Corretto JDK 21" \
      org.opencontainers.image.description="Amazon Corretto OpenJDK distribution"

USER root

# Install Corretto JDK in single layer with modern GPG handling
RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt/lists \
    set -eux; \
    mkdir -p /etc/apt/keyrings \
 && curl -fsSL https://apt.corretto.aws/corretto.key | gpg --dearmor > /etc/apt/keyrings/corretto.gpg \
 && echo "deb [signed-by=/etc/apt/keyrings/corretto.gpg] https://apt.corretto.aws stable main" > /etc/apt/sources.list.d/corretto.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends \
        java-21-amazon-corretto-jdk \
 && rm -rf /tmp/* /var/tmp/*

# Set JAVA_HOME and PATH
ENV JAVA_HOME=/usr/lib/jvm/java-21-amazon-corretto
ENV PATH="/usr/lib/jvm/java-21-amazon-corretto/bin:${PATH}"

USER docker
