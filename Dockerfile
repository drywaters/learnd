# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

# Install build tools
RUN apk add --no-cache make curl && \
    go install github.com/a-h/templ/cmd/templ@latest

# Install standalone Tailwind CSS binary (faster than npm, no Node required)
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-arm64 && \
    chmod +x tailwindcss-linux-arm64 && \
    mv tailwindcss-linux-arm64 /usr/local/bin/tailwindcss

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate templ files and build Tailwind CSS
RUN make templ tail-prod

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/learnd ./cmd/learnd

FROM alpine:3.20
ARG LOG_LEVEL=info
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/learnd ./learnd
COPY --from=builder /src/static ./static
COPY migrations ./migrations

RUN addgroup -S learnd \
    && adduser -S -G learnd learnd \
    && chown -R learnd:learnd /app

ENV PORT=4500
ENV LOG_LEVEL=${LOG_LEVEL}
USER learnd

EXPOSE 4500
CMD ["./learnd"]
