# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/learnd ./cmd/learnd

FROM alpine:3.20
ARG LOG_LEVEL=info
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/learnd ./learnd
COPY migrations ./migrations
COPY static ./static

RUN addgroup -S learnd \
    && adduser -S -G learnd learnd \
    && chown -R learnd:learnd /app

ENV PORT=4500
ENV LOG_LEVEL=${LOG_LEVEL}
USER learnd

EXPOSE 4500
CMD ["./learnd"]
