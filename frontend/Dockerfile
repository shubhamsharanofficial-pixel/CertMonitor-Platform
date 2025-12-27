# --- Stage 1: Builder ---
# Updated to match your local version (v25)
FROM node:25-alpine AS builder

WORKDIR /app

# Copy package files first to leverage caching
COPY package.json package-lock.json ./
RUN npm ci

# Copy the rest of the source code
COPY . .

# Build the React app for production
# This creates a /app/dist folder with static HTML/JS/CSS
RUN npm run build

# --- Stage 2: Production Web Server ---
FROM nginx:alpine

# Remove default Nginx website
RUN rm -rf /usr/share/nginx/html/*

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy our custom Nginx config to replace the default
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80 to the host
EXPOSE 80

# Run Nginx in foreground
CMD ["nginx", "-g", "daemon off;"]