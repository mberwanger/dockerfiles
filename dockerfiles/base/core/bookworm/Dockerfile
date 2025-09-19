# GENERATED FILE, DO NOT MODIFY!
# To update this file please edit the relevant source file and run the generation
# task 'go run cmd/main.go -generate core'
ARG REGISTRY=ghcr.io/mberwanger
FROM ${REGISTRY}/tini:v0.19.0 AS tini
FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        locales \
        gnupg2 \
        adduser \
 && echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen \
 && locale-gen en_US.UTF-8 \
 && dpkg-reconfigure --frontend=noninteractive locales \
 && rm -rf /tmp/* /var/tmp/ \
 && addgroup --gid 9999 docker \
 && adduser --uid 9999 --gid 9999 --disabled-password --system --gecos "Docker User" docker \
 && usermod -L docker \
 && mkdir -p /home/docker \
 && chown docker:docker /home/docker

# Install certificates
# COPY certs/*.crt /usr/local/share/ca-certificates/
# RUN update-ca-certificates \
#  && c_rehash /etc/ssl/certs

# Install Tini for init use (reaps defunct processes and forwards signals)
COPY --from=tini /tini /tini
RUN chmod +x /tini

# Switch to the 'docker' user
USER docker
WORKDIR /home/docker
