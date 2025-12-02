import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  root: 'src',
  base: './',
  build: {
    outDir: '../dist',
    emptyOutDir: true,
  },
});
