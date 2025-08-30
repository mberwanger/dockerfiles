# GENERATED FILE, DO NOT MODIFY!
#
# To update this file please edit the relevant template file and run:
#   go run tool/main.go generate image core
#
# Or regenerate all images with:
#   go run tool/main.go generate all

ARG REGISTRY=ghcr.io/mberwanger
FROM ${REGISTRY}/tini:v0.19.0 AS tini

FROM ubuntu:focal

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=en_US.UTF-8 \
    LC_ALL=en_US.UTF-8

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
        adduser \
        ca-certificates \
        curl \
        gnupg2 \
        locales \
        zip \
 && echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen \
 && locale-gen en_US.UTF-8 \
 && dpkg-reconfigure --frontend=noninteractive locales \
 && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/ \
 && addgroup --gid 9999 docker \
 && adduser --uid 9999 --gid 9999 --disabled-password --gecos "Docker User" docker \
 && install -d -m 700 -o docker -g docker /home/docker

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
