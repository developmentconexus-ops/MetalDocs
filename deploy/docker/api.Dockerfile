FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOFLAGS=-mod=mod go build -o /out/metaldocs-api ./apps/api/cmd/metaldocs-api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/metaldocs-api /app/metaldocs-api
COPY db/migrations /app/db/migrations
EXPOSE 8081
ENTRYPOINT ["/app/metaldocs-api"]
