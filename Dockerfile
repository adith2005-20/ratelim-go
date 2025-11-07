# ---- Build stage ----
FROM golang:1.24.3 AS builder
WORKDIR /app

# Copy go.mod and go.sum first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build a statically linked binary for Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ratelim ./server

# ---- Runtime stage ----
FROM gcr.io/distroless/base-debian12
WORKDIR /app

COPY --from=builder /app/ratelim .
COPY config.yaml .

EXPOSE 8080
CMD ["./ratelim"]
