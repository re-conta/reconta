import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";

// Certificado local gerado por `make certs` (ver Makefile / README) para
// servir o dev server em https://reconta.local. Só é usado quando o `make
// dev` define VITE_HTTPS=1 — assim `bun run dev` isolado continua em HTTP
// em localhost:5173 mesmo que o certificado já exista em disco.
const certDir = resolve(__dirname, "../certs");
const certFile = resolve(certDir, "reconta.local.pem");
const keyFile = resolve(certDir, "reconta.local-key.pem");
const localHttps =
  process.env.VITE_HTTPS === "1" && existsSync(certFile) && existsSync(keyFile)
    ? { cert: readFileSync(certFile), key: readFileSync(keyFile) }
    : undefined;

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  server: {
    https: localHttps,
    allowedHosts: [
      "reconta.app",
      "erebus.paxa.dev",
      "localhost",
      "local.reconta.app",
      "reconta.local",
    ],
    proxy: {
      "/api": {
        target: "http://localhost:3020",
        changeOrigin: true,
      },
    },
  },
});
