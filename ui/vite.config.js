import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
  base: '/static/',
  build: {
    outDir: resolve(__dirname, '../static'),
    emptyOutDir: true,
  },
});
