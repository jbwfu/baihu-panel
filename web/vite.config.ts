import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8052',
        changeOrigin: true,
        ws: true
      }
    }
  },
  build: {
    // 使用相对路径，这样可以部署在任何路径下
    // 资源引用会使用 ./assets/ 而不是 /assets/
    rollupOptions: {
      output: {
        // 确保资源使用相对路径
        assetFileNames: 'assets/[name]-[hash][extname]',
        chunkFileNames: 'assets/[name]-[hash].js',
        entryFileNames: 'assets/[name]-[hash].js'
      }
    }
  },
  // 使用相对路径作为 base，这样资源会相对于 HTML 文件加载
  base: './'
})
