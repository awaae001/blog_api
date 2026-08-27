import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import path from 'path'

export default defineConfig({
  base: '/panel/',
  plugins: [
    vue(),
    Components({
      dts: 'src/components.d.ts',
      resolvers: [ElementPlusResolver({ importStyle: 'css' })]
    })
  ],
  resolve: {
    alias: {
      '@': path.resolve(import.meta.dirname, './src')
    }
  },
  server: {
    host: true,
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:10024',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: '../data/panel',
    emptyOutDir: true,
    rolldownOptions: {
      output: {
        codeSplitting: {
          groups: [
            {
              name: 'element-plus',
              test: /node_modules\/(?:\.pnpm\/)?(?:element-plus|@element-plus)/
            },
            {
              name: 'echarts',
              test: /node_modules\/(?:\.pnpm\/)?(?:echarts|zrender)/
            },
            {
              name: 'editor',
              test: /node_modules\/(?:\.pnpm\/)?(?:@codemirror|codemirror|@lezer)/
            }
          ]
        }
      }
    }
  }
})
