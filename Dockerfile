# Multi-stage build for Lolicount.
# Stage 1: build the Nuxt SSG frontend (Node 22 + pnpm 11).
# Stage 2: build the Go backend (themes + dist embedded-ready, pure Go, no CGO).
# Stage 3: minimal alpine runtime with the binary and data volume.

# ---- Stage 1: frontend ----
FROM node:22-alpine AS frontend
WORKDIR /app
RUN corepack enable
COPY web/pnpm-lock.yaml web/package.json web/.npmrc ./web/
RUN cd web && pnpm install --config.dangerouslyAllowAllBuilds=true
COPY web/ ./web/
RUN cd web && pnpm generate

# ---- Stage 2: backend ----
FROM golang:1.25-alpine AS backend
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy the pre-built frontend dist into assets/dist so go:embed all:dist
# bakes the SSG site into the binary (served by the frontend handler).
COPY --from=frontend /app/web/dist ./assets/dist
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/lolicount ./cmd/server

# ---- Stage 3: runtime ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /out/lolicount /app/lolicount
RUN mkdir -p /app/data
VOLUME ["/app/data"]
EXPOSE 8721
ENV HOST=0.0.0.0 \
    PORT=8721 \
    DB_PATH=/app/data/count.db
ENTRYPOINT ["/app/lolicount"]
