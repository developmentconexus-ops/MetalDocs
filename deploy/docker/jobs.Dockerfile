FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOFLAGS=-mod=mod go build -o /out/metaldocs-jobs ./apps/jobs/cmd/metaldocs-jobs

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/metaldocs-jobs /app/metaldocs-jobs
ENTRYPOINT ["/app/metaldocs-jobs"]
