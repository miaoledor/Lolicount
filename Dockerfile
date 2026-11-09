# Multi-stage build for Lolicount.
# Stage 1: build the Nuxt SSG frontend (Node 22 + pnpm 11).
# Stage 2: build the Go backend (themes + dist embedded-ready, pure Go, no CGO).
# Stage 3: minimal alpine runtime with the binary and data volume.

# ---- Stage 1: frontend ----
# BASE_URL is baked into the SSG bundle at build time (runtimeConfig.public.
# baseUrl → window.__NUXT__.config.public.baseUrl), so the embed links shown
# on the web UI point at the public domain. APP_VERSION carries the release
# tag shown in the footer (runtimeConfig.public.appVersion); CD passes the
# tag as a build arg so no source file needs editing on release. Pass via:
#   docker build --build-arg BASE_URL=https://lolicount.top \
#                --build-arg APP_VERSION=v0.4.4 -t lolicount .
# Empty = embed links fall back to same-origin; the version is omitted.
FROM node:22-alpine AS frontend
ARG BASE_URL=""
ARG APP_VERSION=""
ENV NUXT_PUBLIC_BASE_URL=$BASE_URL
ENV NUXT_PUBLIC_APP_VERSION=$APP_VERSION
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY web/package.json ./web/
RUN corepack enable && corepack install && pnpm --filter web install --frozen-lockfile
COPY web/ ./web/
# Copy the whole scripts tree so test:frontend and the Nuxt config can keep
# resolving their helpers as new checks are added. The localized READMEs are
# test inputs for the repository statistics card consistency check.
COPY scripts/ ./scripts/
COPY README.md README.zh.md README.ja.md ./
RUN pnpm test:frontend && pnpm --dir web generate

# ---- Stage 2: backend ----
FROM golang:1.25-alpine AS backend
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy the pre-built frontend dist into assets/dist so go:embed all:dist
# bakes the SSG site into the binary (served by the frontend handler).
COPY --from=frontend /app/web/.output/public ./assets/dist
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/lolicount ./cmd/server

# ---- Stage 3: runtime ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /out/lolicount /app/lolicount
# PSB emote models are NOT embedded in the binary — they are streamed from
# disk per request (PSB_DIR), so the runtime image carries the directory.
COPY --from=backend /src/assets/psb /app/assets/psb
RUN mkdir -p /app/data
VOLUME ["/app/data"]
EXPOSE 9721
ENV HOST=0.0.0.0 \
    PORT=9721 \
    PSB_DIR=/app/assets/psb \
    DB_PATH=/app/data/count.db
ENTRYPOINT ["/app/lolicount"]
