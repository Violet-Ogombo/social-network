### Backend Dockerfile (multi-stage)
FROM golang:1.21-alpine AS builder
WORKDIR /src

# Install sqlite3 client for applying migrations during build
RUN apk add --no-cache sqlite sqlite-dev bash

# Copy go modules and download
COPY go.mod go.sum /src/
RUN go mod download

# Copy backend code
COPY backend/ /src/backend/

# Build the binary
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/social-network ./backend

# Copy migrations into the image; migrations will be applied at container startup
COPY backend/db/migrations/sqlite /migrations

### Final image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates sqlite bash
COPY --from=builder /usr/local/bin/social-network /usr/local/bin/social-network
COPY --from=builder /src/backend /src/backend
COPY --from=builder /migrations /migrations

# Copy entrypoint (added below) and make it executable
COPY backend/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

WORKDIR /app
ENV DB_PATH=/app/backend/socialnetwork.db
EXPOSE 8080
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/usr/local/bin/social-network"]
