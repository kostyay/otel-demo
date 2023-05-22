# Stage 2: Create the final container with distroless base image (Alpine variant)
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy the built Go binary from the previous stage
COPY app .

# Set the entrypoint
ENTRYPOINT ["./app"]
