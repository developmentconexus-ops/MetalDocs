# TRANSITIONAL local maximum under this repository's "labelled or it's a
# defect" rule (CLAUDE.md, "Global Maximum, Not Local Maximum"): Docker's
# FROM instruction resolves its image reference at parse time, from a
# literal or an ARG supplied via --build-arg -- it cannot read go.mod, so
# this version cannot be derived here, only guarded. Guarded by tools/verify
# check "dockerfile-go-version-drift" (scripts/check-dockerfile-go-version.sh),
# which fails CI when this line's Go version is below go.mod's `go` directive
# (a higher builder version is allowed).
# Global maximum: one parameterized multi-stage Dockerfile shared by
# api/jobs/worker with GO_VERSION threaded from go.mod through the compose
# build stanza (planned A7.4/A7.5) deletes this restatement outright; that
# consolidation is deliberately out of scope here.
FROM golang:1.26.5-alpine AS builder
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
