import tailwindcss from "@tailwindcss/vite";
import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";

// import path from "path";

export default defineConfig({
  base: '/',
  plugins: [tailwindcss(), sveltekit()],
  resolve: {
    // $lib: path.resolve("./src/lib"),
  },
});
