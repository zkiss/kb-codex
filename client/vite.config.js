import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { resolve } from 'path';

export default defineConfig(({ command }) => {
  const isBuild = command === 'build';
  return {
    plugins: [react()],
    server: {
      proxy: {
        '/api': 'http://localhost:8080',
      },
    },
    base: isBuild ? '/static/' : '/',
    build: isBuild
      ? {
          outDir: resolve(__dirname, '../server/static'),
          emptyOutDir: true,
        }
      : {},
    test: {
      globals: true,
      environment: 'jsdom',
      setupFiles: './setupTests.js'
    },
  };
});
