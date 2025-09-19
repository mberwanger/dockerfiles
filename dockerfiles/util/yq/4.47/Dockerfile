# GENERATED FILE, DO NOT MODIFY!
# To update this file please edit the relevant source file and run the generation
# task 'go run cmd/main.go -generate yq'
FROM alpine:3.22

ENV YQ_VERSION=4.47

RUN apk update && \
    apk add --no-cache "yq~${YQ_VERSION}"

ENTRYPOINT ["yq"]
