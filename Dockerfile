# Stage 1: Build the Go binary
FROM golang:1.18 as builder
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o out ./cmd

# Stage 2: Copy the binary into a minimal image
FROM alpine:latest
WORKDIR /root/

# Install necessary CA certificates for HTTPS if required
RUN apk --no-cache add ca-certificates

# Copy the static Go binary from the builder stage
COPY --from=builder /app/out .

# Run the binary
CMD ["./out"]

