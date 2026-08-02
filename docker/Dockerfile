# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build tools
RUN apk add --no-cache git gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o aurix ./cmd/server

# Final lightweight production image
FROM alpine:3.19

WORKDIR /app

# Install FFmpeg runtime dependency for audio decoding
RUN apk add --no-cache ffmpeg ca-certificates tzdata

COPY --from=builder /app/aurix /app/aurix

EXPOSE 2333

ENTRYPOINT ["/app/aurix"]
