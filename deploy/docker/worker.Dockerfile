FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOFLAGS=-mod=mod go build -o /out/metaldocs-worker ./apps/worker/cmd/metaldocs-worker

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
RUN addgroup -g 10001 -S metaldocs && adduser -u 10001 -S -G metaldocs -H -s /sbin/nologin metaldocs
WORKDIR /app
COPY --from=builder --chown=metaldocs:metaldocs /out/metaldocs-worker /app/metaldocs-worker
USER metaldocs
ENTRYPOINT ["/app/metaldocs-worker"]
