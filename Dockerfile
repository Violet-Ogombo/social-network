################################################################################
# Multi-stage Dockerfile
# Stage 1: build frontend (Vite)
# Stage 2: build backend (Go), copy frontend build into backend static dir
# Stage 3: final minimal image with compiled binary and pre-migrated SQLite DB
################################################################################

### Frontend build stage
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend

# Install dependencies and build static assets
COPY frontend/package*.json ./
RUN apk add --no-cache python3 make g++ || true
COPY frontend/ ./
RUN npm ci --silent && npm run build


### Backend build & migrations stage
FROM golang:1.21-alpine AS backend-builder
WORKDIR /app

# Install sqlite3 CLI so we can run migrations during image build
RUN apk add --no-cache sqlite sqlite-dev bash

# Copy backend source
COPY backend/ ./backend/
# Copy go.mod / go.sum from repo root if present
COPY go.mod go.sum ./
RUN go mod download

# Copy frontend dist from frontend-builder into backend so backend can serve it
COPY --from=frontend-builder /app/frontend/dist ./backend/frontend_dist

# Build the Go binary
RUN CGO_ENABLED=0 go build -o /app/social-network ./backend

# Create and apply migrations one-by-one to produce a pre-migrated SQLite DB
RUN mkdir -p /app/backend && touch /app/backend/socialnetwork.db
COPY backend/db/migrations/sqlite /app/migrations
RUN for f in $(ls -1 /app/migrations/*up.sql 2>/dev/null | sort); do echo "Applying migration: $f"; sqlite3 /app/backend/socialnetwork.db < "$f" || (echo "Migration failed: $f"; exit 1); done


### Final runtime image
FROM alpine:3.18
WORKDIR /app

# Add CA certs and sqlite runtime in case it's needed by the app
RUN apk add --no-cache ca-certificates sqlite

# Copy binary and pre-built data
COPY --from=backend-builder /app/social-network /app/social-network
COPY --from=backend-builder /app/backend/socialnetwork.db /app/backend/socialnetwork.db
COPY --from=backend-builder /app/backend/frontend_dist /app/frontend_dist

ENV DB_PATH=/app/backend/socialnetwork.db

EXPOSE 8080

CMD ["/app/social-network"]
