import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Leaderboard } from './pages/Leaderboard'

const queryClient = new QueryClient()

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="min-h-screen">
          <header className="pixel-border bg-[var(--card)] pixel-shadow-md mx-4 mt-4 md:mx-auto md:max-w-5xl p-4 flex items-center justify-between">
            <h1 className="pixel-font text-lg md:text-2xl text-[var(--primary)] tracking-wide">
              League of Ren
            </h1>
            <span className="pixel-font text-xs text-[var(--foreground-muted)] tracking-widest uppercase">
              Translation Rankings
            </span>
          </header>
          <main className="mx-4 md:mx-auto md:max-w-5xl py-6">
            <Routes>
              <Route path="/" element={<Leaderboard />} />
            </Routes>
          </main>
          <footer className="bg-[var(--foreground)] text-[var(--background)] py-8 mt-16">
            <div className="mx-auto max-w-5xl px-6 text-center">
              <p className="pixel-font text-xs tracking-wide">Made with ♥ and pixels</p>
            </div>
          </footer>
        </div>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
