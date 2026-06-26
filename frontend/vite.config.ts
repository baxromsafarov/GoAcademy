import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  // Same-origin in dev: proxy API + static to the backend so the browser never
  // makes a cross-origin request (no CORS needed). In production nginx does the
  // same (see frontend/nginx.conf). The API client therefore uses a relative
  // base ("/api/v1").
  server: {
    // Listen on all interfaces so other devices on the LAN can open the app at
    // http://<this-machine-ip>:5173. The proxy still reaches the backend on the
    // host's own localhost, so only this port needs to be exposed.
    host: true,
    // Allow the dev server to answer requests proxied through a tunnel
    // (cloudflared / ngrok) so the app can be presented from anywhere.
    allowedHosts: true,
    proxy: {
      '/api': 'http://localhost:8080',
      '/static': 'http://localhost:8080',
    },
  },
  // `vite preview` (serving the production build) proxies the same way, so the
  // built frontend can run against the backend without CORS in a local deploy.
  preview: {
    host: true,
    proxy: {
      '/api': 'http://localhost:8080',
      '/static': 'http://localhost:8080',
    },
  },
})
