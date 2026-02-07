import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App'

declare const __COMMIT_HASH__: string
declare const __COMMIT_DATE__: string

console.log(`[leagueofren] ${__COMMIT_HASH__} (${__COMMIT_DATE__})`)

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
