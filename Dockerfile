# ===========
# Build stage
# ===========
FROM golang:1.24 AS builder

# Install build dependencies and Node.js
RUN apt-get update && apt-get install -y wget ca-certificates curl && rm -rf /var/lib/apt/lists/*
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && rm -rf /var/lib/apt/lists/*

# Install templ CLI
RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /app

# Copy go mod files and download Go deps
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# (Tailwind v4) Install Tailwind core + CLI into the builder image
# This creates node_modules in /app, used only in this stage
RUN npm install --save-dev tailwindcss@latest @tailwindcss/cli@latest

# Generate templ files
RUN templ generate

# Build CSS with Tailwind v4 CLI
# Ensure web/input.css uses:  @import "tailwindcss";
RUN npx @tailwindcss/cli@latest -i ./web/input.css -o ./public/style.css --minify

# Build the Go application
RUN CGO_ENABLED=1 GOOS=linux go build -o watchma ./cmd/main.go


# ============
# Runtime stage
# ============
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary and assets from the builder
COPY --from=builder /app/watchma .
COPY --from=builder /app/public ./public

EXPOSE 58008

CMD ["./watchma"]

