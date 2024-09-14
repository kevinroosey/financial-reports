# Use an official Go image to build the binary
FROM golang:1.18 as builder
WORKDIR /app

# Copy go.mod and go.sum to the workspace
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN go build -o out ./cmd

# Use a lightweight image to run the binary
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/out .

# Run the binary
CMD ["./out"]
