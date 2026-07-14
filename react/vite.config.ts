import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { resolve } from "path";

// Multi-page: session-page + popup apps.
export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      input: {
        "session-page": resolve(__dirname, "session-page.html"),
        popup: resolve(__dirname, "popup.html"),
      },
    },
  },
});
