# Build stage
FROM node:alpine AS node-builder

# Install tailwindcss
RUN npm install -g tailwindcss

FROM golang:1.23.1-alpine AS builder

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@v0.2.793

# Install build dependencies
RUN apk add --no-cache make nodejs npm

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install tailwindcss
RUN npm install -g tailwindcss

# Copy source code
COPY . .

# Run the build
RUN make build

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .
# Copy any static assets needed for runtime
COPY --from=builder /app/cmd/web/assets/css/output.css ./cmd/web/assets/css/output.css

# Expose the port your app runs on
EXPOSE 8080

# Set environment variable to enable Go stack traces
ENV GOTRACEBACK=all

# Run the binary
CMD ["./main"]
