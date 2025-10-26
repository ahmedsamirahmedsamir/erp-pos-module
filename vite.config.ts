import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

export default defineConfig({
  plugins: [react()],
  build: {
    lib: {
      entry: resolve(__dirname, 'components/index.ts'),
      name: 'POSModule',
      formats: ['es'],
      fileName: 'pos'
    },
    rollupOptions: {
      external: ['react', 'react-dom', 'react/jsx-runtime', '@erp-modules/shared'],
      output: {
        globals: {
          react: 'React',
          'react-dom': 'ReactDOM',
          '@erp-modules/shared': 'ERPModulesShared'
        }
      }
    },
    sourcemap: true,
    outDir: 'dist'
  }
})

