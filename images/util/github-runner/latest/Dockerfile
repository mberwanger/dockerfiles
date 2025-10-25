# GENERATED FILE, DO NOT MODIFY!
#
# To update this file please edit the relevant template file and run:
#   go run tool/main.go generate image github-runner
#
# Or regenerate all images with:
#   go run tool/main.go generate all

ARG REGISTRY=ghcr.io/mberwanger
FROM ${REGISTRY}/core:noble AS build

ARG TARGETARCH
ARG TARGETOS
ARG DOCKER_VERSION=28.4.0
ARG BUILDX_VERSION=0.29.0
ARG RUNNER_VERSION=2.328.0
ARG RUNNER_CONTAINER_HOOKS_VERSION=0.7.0

USER root
WORKDIR /actions-runner

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
        unzip \
 && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Download all assets in one layer
RUN set -ex \
    && RUNNER_ARCH=${TARGETARCH} \
    && DOCKER_ARCH=${TARGETARCH} \
    # Convert architectures
    && if [ "${TARGETARCH}" = "amd64" ]; then \
         RUNNER_ARCH=x64; \
         DOCKER_ARCH=x86_64; \
       elif [ "${TARGETARCH}" = "arm64" ]; then \
         DOCKER_ARCH=aarch64; \
       fi \
    # Download GitHub Actions runner
    && curl -fL --retry 3 -o runner.tar.gz \
         "https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-${TARGETOS}-${RUNNER_ARCH}-${RUNNER_VERSION}.tar.gz" \
    && tar xzf runner.tar.gz \
    && rm runner.tar.gz \
    # Download runner hooks
    && curl -fL --retry 3 -o runner-container-hooks.zip \
         "https://github.com/actions/runner-container-hooks/releases/download/v${RUNNER_CONTAINER_HOOKS_VERSION}/actions-runner-hooks-k8s-${RUNNER_CONTAINER_HOOKS_VERSION}.zip" \
    && unzip -q runner-container-hooks.zip -d ./k8s \
    && rm runner-container-hooks.zip \
    # Download Docker
    && curl -fL --retry 3 -o docker.tgz \
         "https://download.docker.com/${TARGETOS}/static/stable/${DOCKER_ARCH}/docker-${DOCKER_VERSION}.tgz" \
    && tar xzf docker.tgz \
    && rm docker.tgz \
    # Download buildx plugin
    && mkdir -p /usr/local/lib/docker/cli-plugins \
    && curl -fL --retry 3 -o /usr/local/lib/docker/cli-plugins/docker-buildx \
         "https://github.com/docker/buildx/releases/download/v${BUILDX_VERSION}/buildx-v${BUILDX_VERSION}.linux-${TARGETARCH}" \
    && chmod +x /usr/local/lib/docker/cli-plugins/docker-buildx

FROM ${REGISTRY}/core:noble

ENV RUNNER_MANUALLY_TRAP_SIG=1
ENV ACTIONS_RUNNER_PRINT_LOG_TO_STDOUT=1

USER root

RUN curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor > /usr/share/keyrings/cloud.google.gpg \
 && echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" > /etc/apt/sources.list.d/google-cloud-sdk.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends \
        build-essential \
        git \
        git-lfs \
        google-cloud-cli \
        jq \
        libicu74 \
        lsb-release \
        nodejs \
        npm \
        python3 \
        python3-dev \
        python3-pip \
        python3-setuptools \
        python3-venv \
        python3-wheel \
        ruby-full \
        software-properties-common \
        sudo \
        unzip \
        wget \
        yq \
        zip \
 && chmod -R 777 /opt \
 && chmod -R 777 /usr/share \
 && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN KUBECTL_MINOR_VERSION=$(curl -fsSL --retry 5 --retry-delay 10 "https://dl.k8s.io/release/stable.txt" | cut -d'.' -f1,2) \
 && curl -fsSL --retry 5 --retry-delay 10 "https://pkgs.k8s.io/core:/stable:/${KUBECTL_MINOR_VERSION}/deb/Release.key" | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg \
 && echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/${KUBECTL_MINOR_VERSION}/deb/ /" > /etc/apt/sources.list.d/kubernetes.list \
 && apt-get update \
 && apt-get install -y kubectl \
 && curl -fsSL https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash \
 && rm -rf /var/lib/apt/lists/*

RUN adduser --disabled-password --gecos "" --uid 1001 runner \
 && usermod -aG sudo runner \
 && usermod -aG docker runner \
 && echo "%sudo   ALL=(ALL:ALL) NOPASSWD:ALL" > /etc/sudoers \
 && echo "Defaults env_keep += \"DEBIAN_FRONTEND\"" >> /etc/sudoers

WORKDIR /home/runner

COPY --chown=runner:docker --from=build /actions-runner .
COPY --from=build /usr/local/lib/docker/cli-plugins/docker-buildx /usr/local/lib/docker/cli-plugins/docker-buildx

RUN install -o root -g root -m 755 docker/* /usr/bin/ && rm -rf docker

USER runner
RUN git config --global user.name "GitHub Actions Runner" \
 && git config --global user.email "actions@github.com" \
 && git config --global --add safe.directory '*'
