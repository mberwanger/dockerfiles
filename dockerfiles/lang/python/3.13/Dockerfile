# GENERATED FILE, DO NOT MODIFY!
# To update this file please edit the relevant source file and run the generation
# task 'go run cmd/main.go -generate python'
ARG REGISTRY=ghcr.io/mberwanger
FROM ${REGISTRY}/core:bullseye

LABEL org.opencontainers.image.authors="Martin Berwanger" \
      org.opencontainers.image.title="Python 3.13.7" \
      org.opencontainers.image.description="Python programming language runtime with pip and pipenv for application development"

USER root

ARG PYTHON_VERSION=3.13.7
ARG PYTHON_MAJOR_MINOR_VERSION=3.13
ARG PIP_VERSION=25.2
ARG PGP_KEY_ID=A821E680E5FA6305

# Install Python build dependencies and compile from source
RUN set -eux; \
    apt-get update \
 && apt-get install -y --no-install-recommends \
        build-essential \
        libbz2-dev \
        libffi-dev \
        libgdbm-dev \
        liblzma-dev \
        libncurses5-dev \
        libnss3-dev \
        libreadline-dev \
        libsqlite3-dev \
        libssl-dev \
        xz-utils \
        zlib1g-dev \
 && cd /tmp \
 && curl -fsSLO https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tar.xz \
 && curl -fsSLO https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tar.xz.asc \
 && export GNUPGHOME="$(mktemp -d)" \
 && gpg --batch --keyserver hkps://keys.openpgp.org --recv-keys ${PGP_KEY_ID} || \
    gpg --batch --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys ${PGP_KEY_ID} \
 && gpg --verify Python-${PYTHON_VERSION}.tar.xz.asc Python-${PYTHON_VERSION}.tar.xz \
 && gpgconf --kill all \
 && rm -rf "$GNUPGHOME" \
 && tar -xf Python-${PYTHON_VERSION}.tar.xz \
 && cd Python-${PYTHON_VERSION} \
 && ./configure --enable-optimizations --with-lto --enable-shared \
 && make -j$(nproc) altinstall \
 && ldconfig \
 && cd / \
 && rm -rf /tmp/Python-${PYTHON_VERSION}* \
 && python${PYTHON_MAJOR_MINOR_VERSION} --version

# Install pip and pipenv
RUN pip${PYTHON_MAJOR_MINOR_VERSION} install --no-cache-dir --upgrade pip==${PIP_VERSION} \
 && pip${PYTHON_MAJOR_MINOR_VERSION} install --no-cache-dir pipenv

# Setup working directory and user
WORKDIR /usr/src/app/
RUN chown docker:docker .

USER docker

# Initialize pipenv with specific Python version
RUN pipenv --python ${PYTHON_MAJOR_MINOR_VERSION}

COPY pip-private pipenv-private /usr/local/bin/
