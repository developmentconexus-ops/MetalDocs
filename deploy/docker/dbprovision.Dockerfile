FROM golang:1.26.5-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOFLAGS=-mod=mod go build -o /out/metaldocs-dbprovision ./apps/dbprovision/cmd/metaldocs-dbprovision

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
RUN addgroup -g 10001 -S metaldocs && adduser -u 10001 -S -G metaldocs -H -s /sbin/nologin metaldocs
WORKDIR /app
COPY --from=builder --chown=metaldocs:metaldocs /out/metaldocs-dbprovision /app/metaldocs-dbprovision
# metaldocs-dbprovision is the one-shot binary that runs BEFORE api/worker/jobs
# (A6.1 re-cut, issue #88): it connects as the bootstrap superuser and applies
# db/prerequisites (extensions), db/grants (identity roles + privilege grants),
# db/migrations, and River's schema, in that order. It is the only image that
# needs these three trees; api/worker/jobs no longer ship them.
COPY --chown=metaldocs:metaldocs db/prerequisites /app/db/prerequisites
COPY --chown=metaldocs:metaldocs db/grants /app/db/grants
COPY --chown=metaldocs:metaldocs db/migrations /app/db/migrations
USER metaldocs
ENTRYPOINT ["/app/metaldocs-dbprovision"]
