# ---------- STAGE 1: BUILD STAGE ----------
# Pin to a specific, stable version of the Go Alpine image.
FROM golang:1.24.4-alpine3.22 AS builder

# Set the working directory.
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker's layer caching.
# This layer only gets invalidated if your dependencies change.
COPY go.mod go.sum ./

# Download the dependencies.
RUN go mod download

# Now, copy the rest of your application source code.
# This layer will be invalidated if your source code changes, but the
# dependency layer above will remain cached.
COPY . .

# Build the Go application as a static binary.
# - We assume the main package is in `.` for consistency with the debug stage.
# - Using -ldflags "-w -s" strips debug information, reducing binary size.
# - GOOS=linux and GOARCH=amd64 ensure the binary is built for the target OS.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/sidecar .

# ---------- STAGE 2: DEBUGGING STAGE (Optional for Local Dev) ----------
# Use the same pinned Go base image for consistency.
FROM golang:1.24.4-alpine3.22 AS debug

# Set the working directory
WORKDIR /app

# Copy dependency files and download them.
COPY go.mod go.sum ./
RUN go mod download

# Install CompileDaemon for hot-reloading.
RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon

# Copy the rest of the application files.
COPY . .

# Entry point for the debugging container.
# Watches for file changes and rebuilds/reruns the app.
ENTRYPOINT ["CompileDaemon", "-build=go build -o /build/app .", "-command=/build/app", "-polling", "-graceful-kill"]

# ---------- STAGE 3: DEPLOYMENT STAGE ----------
# Use a specific, minimal base image. `scratch` is the absolute smallest,
# but `alpine` is often better as it provides a shell for debugging if needed.
FROM alpine:3.20 AS deploy

# Create a non-root user and group for the application to run as.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory.
WORKDIR /app

# Copy the statically compiled Go binary from the builder stage.
# Also copy ownership information for the new user.
COPY --from=builder --chown=appuser:appgroup /app/sidecar .

# Switch to the non-root user.
USER appuser

# Default command to run the application.
CMD ["./sidecar"]