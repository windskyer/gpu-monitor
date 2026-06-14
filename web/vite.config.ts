import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  base: '/gpu/',
  plugins: [vue()],

  server: {
    port: 5173,
    proxy: {
      // dev 时把 WS 和 API 代理到 Go 后端
      '/ws':  { target: 'ws://localhost:8800',   ws: true,         changeOrigin: true },
      '/api': { target: 'http://localhost:8800',                   changeOrigin: true },
    },
  },

  build: {
    // 产物直接输出到 Go embed 目录，打包进二进制
    outDir:    '../internal/server/dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        // 固定文件名，避免 hash 让 embed 文件列表不可预测
        entryFileNames: 'assets/[name].js',
        chunkFileNames: 'assets/[name].js',
        assetFileNames: 'assets/[name].[ext]',
      },
    },
  },
})
