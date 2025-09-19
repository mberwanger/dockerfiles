# GENERATED FILE, DO NOT MODIFY!
# To update this file please edit the relevant source file and run the generation
# task 'go run cmd/main.go -generate python'
ARG REGISTRY=ghcr.io/mberwanger
FROM ${REGISTRY}/core:bullseye

LABEL org.opencontainers.image.authors="Martin Berwanger" \
      org.opencontainers.image.title="Python 3.11.13" \
      org.opencontainers.image.description="Python programming language runtime with pip and pipenv for application development"

USER root

ARG PYTHON_VERSION=3.11.13
ARG PYTHON_MAJOR_MINOR_VERSION=3.11
ARG PIP_VERSION=24.3.1
ARG PYTHON_GPG_KEY=64E628F8D684696D

# Install build dependencies for Python compilation
RUN set -eux; \
    apt-get update \
 && apt-get install -y --no-install-recommends \
        build-essential \
        wget \
        ca-certificates \
        gnupg \
        dirmngr \
        libssl-dev \
        zlib1g-dev \
        libbz2-dev \
        libreadline-dev \
        libsqlite3-dev \
        libncursesw5-dev \
        xz-utils \
        tk-dev \
        libxml2-dev \
        libxmlsec1-dev \
        libffi-dev \
        liblzma-dev

# Download, verify, and compile Python from source
RUN set -eux; \
    wget -O python.tgz "https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tgz" \
 && wget -O python.tgz.asc "https://www.python.org/ftp/python/${PYTHON_VERSION}/Python-${PYTHON_VERSION}.tgz.asc" \
 && export GNUPGHOME="$(mktemp -d)" \
 && gpg --batch --keyserver hkps://keys.openpgp.org --recv-keys "${PYTHON_GPG_KEY}" \
 && gpg --batch --verify python.tgz.asc python.tgz \
 && { command -v gpgconf > /dev/null && gpgconf --kill all || :; } \
 && rm -rf "$GNUPGHOME" python.tgz.asc \
 && mkdir -p /usr/src/python \
 && tar -xzC /usr/src/python --strip-components=1 -f python.tgz \
 && rm python.tgz \
 && cd /usr/src/python \
 && gnuArch="$(dpkg-architecture --query DEB_BUILD_GNU_TYPE)" \
 && ./configure \
        --build="$gnuArch" \
        --enable-loadable-sqlite-extensions \
        --enable-optimizations \
        --enable-option-checking=fatal \
        --enable-shared \
        --with-system-expat \
        --without-ensurepip \
 && make -j "$(nproc)" \
 && make install \
 && rm -rf /usr/src/python \
 && python${PYTHON_MAJOR_MINOR_VERSION} --version \
 && find /usr/local -depth \
        \( \
            \( -type d -a \( -name test -o -name tests -o -name idle_test \) \) \
            -o \( -type f -a \( -name '*.pyc' -o -name '*.pyo' -o -name 'libpython*.a' \) \) \
        \) -exec rm -rf '{}' + \
 && ldconfig

# Install pip manually (since we compiled without ensurepip)
RUN set -eux; \
    wget -O get-pip.py https://bootstrap.pypa.io/get-pip.py \
 && python${PYTHON_MAJOR_MINOR_VERSION} get-pip.py --no-cache-dir "pip==${PIP_VERSION}" \
 && rm get-pip.py \
 && python${PYTHON_MAJOR_MINOR_VERSION} -m pip install --no-cache-dir pipenv \
 && ln -sf python${PYTHON_MAJOR_MINOR_VERSION} /usr/bin/python3 \
 && ln -sf python${PYTHON_MAJOR_MINOR_VERSION} /usr/bin/python

# Setup working directory and user
WORKDIR /usr/src/app/
RUN chown docker:docker .

USER docker

# Initialize pipenv with specific Python version
RUN pipenv --python ${PYTHON_MAJOR_MINOR_VERSION}
