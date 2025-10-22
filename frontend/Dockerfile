### Frontend Dockerfile (build + nginx)
FROM node:18-alpine AS builder
WORKDIR /app
COPY frontend/package*.json ./
RUN apk add --no-cache python3 make g++ || true
COPY frontend/ ./
RUN npm ci --silent && npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
