import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'

// Vite's project root is this directory, so `src/app.js` is the entry point and
// output goes to ../public/build at the repository root.
export default defineConfig({
  plugins: [tailwindcss()],
  build: {
    // Writes .vite/manifest.json, which maps `src/app.js` to the hashed files
    // that were actually produced. The Go side reads it — see assets/.
    manifest: true,
    outDir: '../public/build',
    // Required because outDir is outside this directory.
    emptyOutDir: true,
    rollupOptions: {
      input: 'src/app.js',
    },
  },
})
