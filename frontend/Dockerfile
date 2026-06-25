# --- build the SPA ---
FROM node:22-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
# The API client uses a relative base ("/api/v1") that nginx proxies to the
# backend, so no API URL needs baking in at build time.
RUN npm run build

# --- serve static files + reverse-proxy /api to the backend ---
FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
