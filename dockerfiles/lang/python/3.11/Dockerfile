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

# Install Python from deadsnakes PPA (much faster than compiling)
RUN set -eux; \
    apt-get update \
 && apt-get install -y --no-install-recommends \
        software-properties-common \
        gpg-agent \
 && add-apt-repository ppa:deadsnakes/ppa \
 && apt-get update \
 && apt-get install -y --no-install-recommends \
        python${PYTHON_MAJOR_MINOR_VERSION} \
        python${PYTHON_MAJOR_MINOR_VERSION}-dev \
        python${PYTHON_MAJOR_MINOR_VERSION}-venv \
        python${PYTHON_MAJOR_MINOR_VERSION}-distutils \
 && python${PYTHON_MAJOR_MINOR_VERSION} --version

# Install pip and pipenv
RUN python${PYTHON_MAJOR_MINOR_VERSION} -m pip install --no-cache-dir --upgrade pip==${PIP_VERSION} \
 && python${PYTHON_MAJOR_MINOR_VERSION} -m pip install --no-cache-dir pipenv \
 && ln -sf python${PYTHON_MAJOR_MINOR_VERSION} /usr/bin/python3 \
 && ln -sf python${PYTHON_MAJOR_MINOR_VERSION} /usr/bin/python

# Setup working directory and user
WORKDIR /usr/src/app/
RUN chown docker:docker .

USER docker

# Initialize pipenv with specific Python version
RUN pipenv --python ${PYTHON_MAJOR_MINOR_VERSION}
