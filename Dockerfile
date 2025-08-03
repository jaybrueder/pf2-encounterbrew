# Build stage
FROM golang:1.24.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache make nodejs npm

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@v0.2.793

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install tailwindcss
RUN npm install -g tailwindcss@3.4.17

# Copy source code
COPY . .

# Run the build
RUN make build

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup

# Copy required assets
COPY --from=builder /app/main .
COPY --from=builder /app/cmd/web/assets/css/output.css ./cmd/web/assets/css/output.css
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/data/ ./data/

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose the port your app runs on
EXPOSE 8081

# Set environment variables
ENV PORT=8081
ENV MIGRATIONS_PATH=migrations
ENV DB_HOST=postgres
ENV DB_PORT=5432
ENV DB_DATABASE=encounterbrew
ENV DB_SCHEMA=public
ENV GOTRACEBACK=all

# Note: USERNAME and PASSWORD should be provided at runtime via docker run -e or docker-compose
# for security reasons, not baked into the image

# Run the binary
CMD ["./main"]
