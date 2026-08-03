FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOFLAGS=-mod=mod go build -o /out/metaldocs-api ./apps/api/cmd/metaldocs-api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
RUN addgroup -g 10001 -S metaldocs && adduser -u 10001 -S -G metaldocs -H -s /sbin/nologin metaldocs
WORKDIR /app
COPY --from=builder --chown=metaldocs:metaldocs /out/metaldocs-api /app/metaldocs-api
COPY --chown=metaldocs:metaldocs db/migrations /app/db/migrations
# db/grants is the ledger-less privilege/role stage metaldocs-api re-applies on
# every start (internal/platform/migrate.ApplyGrants) — a missing directory is
# a fatal startup error, so it must ship in the image alongside db/migrations.
COPY --chown=metaldocs:metaldocs db/grants /app/db/grants
EXPOSE 8081
USER metaldocs
ENTRYPOINT ["/app/metaldocs-api"]
