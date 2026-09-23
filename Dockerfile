# ==============================================================================
# Stage 1: Build React 19 Frontend SPA Assets
# ==============================================================================
FROM node:22-alpine AS web-builder
WORKDIR /app/web

# Install frontend dependencies
COPY web/package.json web/package-lock.json ./
RUN npm ci

# Build static bundle to /app/web/dist
COPY web/ ./
RUN npm run build

# ==============================================================================
# Stage 2: Build Static Go Binary
# ==============================================================================
FROM golang:1.24-alpine AS go-builder
WORKDIR /app
RUN apk add --no-cache ca-certificates git

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and compiled static web assets for Go embed.FS
COPY . ./
COPY --from=web-builder /app/web/dist ./web/dist

# Build pure static Linux binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-w -s" \
    -o /app/bin/agy-ge-board \
    ./cmd/agy-ge-board

# ==============================================================================
# Stage 3: Minimal Distroless Production Runtime
# ==============================================================================
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /

# Copy binary and root CA certificates for Google Cloud API TLS
COPY --from=go-builder /app/bin/agy-ge-board /agy-ge-board
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Cloud Run dynamic port binding
ENV PORT=8080
EXPOSE 8080

# Run as non-privileged user
USER nonroot:nonroot

# Default entrypoint allows running CLI subcommands directly (cost, license, doctor, etc.)
ENTRYPOINT ["/agy-ge-board"]

# Default action is to launch the HTTP web server for Cloud Run
CMD ["serve"]
