# Stage 1: Build the Go application
FROM golang:1.20 AS builder

WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/...

# Stage 2: Create the final container with distroless base image (Alpine variant)
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy the built Go binary from the previous stage
COPY --from=builder /app/app .

# Set the entrypoint
ENTRYPOINT ["./app"]
