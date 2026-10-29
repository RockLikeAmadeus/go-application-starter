# syntax=docker/dockerfile:1

# ---- web: build the React SPA ----
FROM node:25-alpine AS web
WORKDIR /src/web
RUN npm install -g pnpm@10
COPY web/package.json web/pnpm-lock.yaml ./
RUN --mount=type=cache,target=/root/.local/share/pnpm/store pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

# ---- go: build the server with the SPA embedded ----
FROM golang:1.27-alpine AS go
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
COPY --from=web /src/web/dist ./web/dist
ARG VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -tags embedweb -trimpath \
      -ldflags "-s -w -X github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli.Version=${VERSION}" \
      -o /out/server ./cmd/server \
 && mkdir -p /out/data

# ---- runtime: distroless (no shell, no package manager) ----
# Runs as root because Fly.io volumes are mounted root-owned and distroless
# has no shell to chown them at startup. Once you move to a managed database
# (no local volume), switch to the ":nonroot" tag and add `USER nonroot:nonroot`.
FROM gcr.io/distroless/static-debian12
COPY --from=go /out/server /server
COPY --from=go /out/data /data
ENV PORT=8080 \
    STARTER_STORAGE=jsonfile \
    STARTER_DATA_DIR=/data \
    STARTER_LOG_FORMAT=json
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/server"]
