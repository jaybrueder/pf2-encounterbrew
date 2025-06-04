# Build stage
FROM golang:1.24.3-alpine AS builder

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

# Set environment variable to enable Go stack traces
ENV PORT=8081
ENV USERNAME=dragon
ENV PASSWORD=hobgoblin
ENV MIGRATIONS_PATH=migrations
ENV DB_HOST=postgres
ENV DB_PORT=5432
ENV DB_DATABASE=encounterbrew
ENV DB_SCHEMA=public
ENV GOTRACEBACK=all

# Run the binary
CMD ["./main"]
