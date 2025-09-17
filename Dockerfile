# GoReleaser builds the Go binary and provides it in the build context.
# This Dockerfile just packages that binary into a minimal image.
# https://goreleaser.com/customization/docker/#the-docker-build-context
FROM gcr.io/distroless/static:nonroot

COPY dynamo-db-todos /
ENTRYPOINT ["/dynamo-db-todos"]
