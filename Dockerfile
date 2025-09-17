# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Final stage
FROM alpine:latest
WORKDIR /app
# Copy the binary from the builder stage and set the owner to the non-root 'nobody' user.
COPY --from=builder --chown=nobody:nobody /app/app .
# Switch to the non-root user.
USER nobody
ENTRYPOINT ["./app"]
