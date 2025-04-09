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
RUN npm install -g tailwindcss@3.4.17

# Copy source code
COPY . .

# Run the build
RUN make build

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy required assets
COPY --from=builder /app/main .
COPY --from=builder /app/cmd/web/assets/css/output.css ./cmd/web/assets/css/output.css
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/data/ ./data/

# Expose the port your app runs on
EXPOSE 8080

# Set environment variable to enable Go stack traces
ENV PORT=8080
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
