# ── Build ─────────────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Download deps first (cached layer unless go.mod/go.sum change)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Build for the target platform. TARGETARCH/TARGETOS are set automatically
# by `docker buildx` when building multi-arch images.
ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build \
      -trimpath \
      -ldflags="-s -w" \
      -o stacksnap \
      .

# ── Runtime ───────────────────────────────────────────────────────────────────
FROM alpine:3.19

# ca-certificates  → HTTPS to Portainer
# docker-cli       → Docker fallback discovery
RUN apk add --no-cache ca-certificates docker-cli

WORKDIR /app
COPY --from=builder /app/stacksnap /usr/local/bin/stacksnap

ENTRYPOINT ["/usr/local/bin/stacksnap"]
