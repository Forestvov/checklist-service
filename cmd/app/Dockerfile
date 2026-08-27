# syntax=docker/dockerfile:1

FROM golang:1.26.4-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY docs ./docs
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/checklist-service \
    ./cmd/app

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S checklist \
    && adduser -S -G checklist -h /app checklist \
    && mkdir -p /app/out/logs \
    && chown -R checklist:checklist /app

WORKDIR /app

COPY --from=builder --chown=checklist:checklist /out/checklist-service ./checklist-service

USER checklist

EXPOSE 5050

ENTRYPOINT ["/app/checklist-service"]
