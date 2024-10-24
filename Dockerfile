FROM golang:1.19-bullseye AS build-base

WORKDIR /smoldb

# Copy only files required to install dependencies (better layer caching)
COPY go.mod go.sum ./

# Use cache mount to speed up install of existing dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go mod download

# Build image
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/smoldb

## Image creation stage
FROM scratch

# Copy user from build stage
COPY --from=builder /etc/passwd /etc/passwd

# Copy smoldb
COPY --from=builder /go/bin/smoldb /go/bin/smoldb
COPY --from=builder /smoldb/src/db /go/bin/db
WORKDIR /go/bin

# Set entrypoint
ENTRYPOINT ["./smoldb"]