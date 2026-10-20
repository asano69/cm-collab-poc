import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

// Vitest bundles its own Vite, which does not understand the
// `resolve.tsconfigPaths` option vite.config.ts uses for the "@/*" alias, so
// the alias is declared explicitly here. Only what the tests need is
// configured; the Solid and Tailwind plugins of vite.config.ts are not
// loaded by any test.
export default defineConfig({
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
    // Solid's default (server) build makes createEffect a no-op; the browser
    // build is needed to test reactive code such as lib/dexie/liveQuery.ts.
    conditions: ["browser"],
  },
  test: {
    environment: "jsdom",
    // Resolve solid-js through Vite (and thus the conditions above) instead of
    // letting Node load its server build.
    server: { deps: { inline: [/solid-js/] } },
  },
});
