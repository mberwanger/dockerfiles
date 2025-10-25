# GENERATED FILE, DO NOT MODIFY!
#
# To update this file please edit the relevant template file and run:
#   go run tool/main.go generate image yq
#
# Or regenerate all images with:
#   go run tool/main.go generate all

FROM alpine:3.22

ARG YQ_VERSION=4.47

RUN apk update && \
    apk add --no-cache "yq~${YQ_VERSION}"

ENTRYPOINT ["yq"]
