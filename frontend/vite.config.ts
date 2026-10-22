import { defineConfig } from "vite";
import solid from "vite-plugin-solid";

const wsProxy: ProxyOptions = {
  target: "http://127.0.0.1:3002",
  changeOrigin: true,
  ws: true,
  configure: (proxy) => {
    proxy.on("proxyReqWs", (proxyReq) => {
      proxyReq.setHeader("origin", "http://127.0.0.1:3002");
    });
  },
};

export default defineConfig({
  plugins: [solid()],
  // Resolves the "@/*" alias declared in tsconfig.json's "paths".
  resolve: {
    tsconfigPaths: true,
  },
  server: {
    host: "0.0.0.0",
    port: 3003,
    allowedHosts: true,
    proxy: {
      // Use 127.0.0.1 explicitly to avoid localhost resolving to ::1 (IPv6)
      // while PocketBase only listens on 127.0.0.1 (IPv4).
      "/api": { target: "http://127.0.0.1:3002", changeOrigin: true },
      "/_": { target: "http://127.0.0.1:3002", changeOrigin: true },
      "/connection/": wsProxy,
    },
  },
  build: {
    outDir: "../internal/static/dist",
    emptyOutDir: true,
  },
});
