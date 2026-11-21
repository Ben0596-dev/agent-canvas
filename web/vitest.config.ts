import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    environment: "node",
    include: ["**/*.{test,spec}.{ts,tsx}"],
  },
  server: {
    host: "0.0.0.0",
    port: 3001,
  },
});
