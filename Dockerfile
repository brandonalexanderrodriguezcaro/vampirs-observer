# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files to the container
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code to the container
COPY . .

# Build the Go application
RUN go build -o observer

# Stage 2: Run the application in a smaller container
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/observer .

# Default command to run the app with the config flag
CMD ["./observer", "-config", "/config/config.conf"]