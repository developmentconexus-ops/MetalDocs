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
# db/migrations and db/grants no longer ship in this image (A6.1 re-cut,
# issue #88): DDL and the privilege/role stage now run once, before this
# container starts, via apps/dbprovision/cmd/metaldocs-dbprovision (see
# deploy/docker/dbprovision.Dockerfile) — metaldocs-api's own connection
# (metaldocs_runtime) holds no DDL rights and never applies these files.
EXPOSE 8081
USER metaldocs
ENTRYPOINT ["/app/metaldocs-api"]
