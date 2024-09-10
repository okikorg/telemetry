# Use ARG to specify the base image
ARG BASE_IMAGE=golang:1.22

# Build stage
FROM ${BASE_IMAGE} AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# ARG for build tags
ARG BUILD_TAGS=""

# Build the application
RUN go build -tags="${BUILD_TAGS}" -o telemetry .

# Final stage
FROM ${BASE_IMAGE}

# Copy the binary from the builder stage
COPY --from=builder /app/telemetry /telemetry

# Expose the port the app runs on
EXPOSE 8080

# Run the binary
CMD ["/telemetry"]
