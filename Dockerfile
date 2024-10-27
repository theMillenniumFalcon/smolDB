FROM golang:1.19-bullseye AS build-base
ENV GO111MODULE=on

WORKDIR /smoldb/src

COPY . ./smoldb/src

# Use cache mount to speed up install of existing dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go mod download

# Build image
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/smoldb

## Image creation stage
FROM scratch

# Copy user from build stage
COPY --from=build-base /etc/passwd /etc/passwd

# Copy smoldb
COPY --from=build-base /go/bin/smoldb /go/bin/smoldb
COPY --from=build-base /smoldb/src/db /go/bin/db
WORKDIR /go/bin

# Set entrypoint
ENTRYPOINT ["./smoldb"]