import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { execSync } from 'child_process'

function git(cmd: string): string {
  try { return execSync(cmd).toString().trim() } catch { return 'unknown' }
}

const commitHash = process.env.VITE_COMMIT_HASH || git('git rev-parse --short HEAD')
const commitDate = process.env.VITE_COMMIT_DATE || git('git log -1 --format=%ci')

export default defineConfig({
  plugins: [react(), tailwindcss()],
  define: {
    __COMMIT_HASH__: JSON.stringify(commitHash),
    __COMMIT_DATE__: JSON.stringify(commitDate),
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:3000',
    },
  },
})
