import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Github, Heart, BookOpen, Swords } from 'lucide-react'
import { Leaderboard } from './pages/Leaderboard'

const queryClient = new QueryClient()

const REPO_URL = 'https://github.com/jusunglee/leagueofren'

function Sparkle({ color = 'var(--accent-bright)' }: { color?: string }) {
  return <span className="pixel-sparkle" style={{ background: color }} aria-hidden="true" />
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="min-h-screen flex flex-col">

          {/* ── Full-width Header ── */}
          <header className="bg-[var(--card)] border-b-4 border-[var(--border)] px-6 py-3">
            <div className="max-w-[1600px] mx-auto flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-[var(--violet)] border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm flex items-center justify-center">
                  <Swords size={20} strokeWidth={2.5} className="text-white" />
                </div>
                <div>
                  <h1 className="pixel-font text-base lg:text-lg text-[var(--violet)] tracking-wide leading-tight">
                    League of Ren
                  </h1>
                  <p className="text-xs text-[var(--foreground-muted)] hidden sm:block">Translation Rankings</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <a
                  href={REPO_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="pixel-font text-xs px-4 py-2 bg-[var(--primary)] text-white border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm tracking-wide uppercase hover:bg-[var(--primary-hover)] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_0px_var(--border)] transition-all duration-150 btn-press hidden sm:inline-flex items-center gap-2 focus-visible:ring-2 focus-visible:ring-[var(--ring)]"
                >
                  Submit Yours ★
                </a>
                <a
                  href={REPO_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="w-10 h-10 bg-[var(--card)] border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm flex items-center justify-center hover:bg-[var(--muted)] hover:-translate-x-0.5 hover:-translate-y-0.5 transition-all duration-150 btn-press"
                  aria-label="GitHub Repository"
                >
                  <Github size={20} strokeWidth={2.5} />
                </a>
              </div>
            </div>
          </header>

          {/* ── Main Layout: Sidebar + Content on desktop, stacked on mobile ── */}
          <div className="flex-1 flex flex-col lg:flex-row max-w-[1600px] mx-auto w-full">

            {/* ── Sidebar (desktop only) ── */}
            <aside className="hidden lg:flex flex-col w-[280px] shrink-0 p-6 gap-6 border-r-4 border-[var(--border)] bg-[var(--card)]">

              {/* About */}
              <div>
                <div className="flex items-center gap-2 mb-3">
                  <div className="w-6 h-6 bg-[var(--sky)] border-4 border-[var(--border)] rounded-[4px] flex items-center justify-center pixel-shadow-sm">
                    <BookOpen size={12} strokeWidth={2.5} className="text-white" />
                  </div>
                  <h3 className="pixel-font text-[10px] tracking-wide">About</h3>
                </div>
                <p className="text-xs text-[var(--foreground-muted)] leading-relaxed">
                  Community-ranked translations of Korean and Chinese League of Legends summoner names.
                  Vote for the best, discover hidden meanings.
                </p>
              </div>

              <div className="border-t-2 border-dashed border-[var(--border-light)]" />

              {/* Contribute */}
              <div>
                <div className="flex items-center gap-2 mb-3">
                  <div className="w-6 h-6 bg-[var(--secondary)] border-4 border-[var(--border)] rounded-[4px] flex items-center justify-center pixel-shadow-sm">
                    <Heart size={12} strokeWidth={2.5} className="text-white" />
                  </div>
                  <h3 className="pixel-font text-[10px] tracking-wide">Contribute</h3>
                </div>
                <p className="text-xs text-[var(--foreground-muted)] leading-relaxed mb-3">
                  Run the bot and opt in to share translations with the community.
                </p>
                <a
                  href={REPO_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="pixel-font text-[10px] px-3 py-1.5 bg-[var(--secondary)] text-white border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm tracking-wide uppercase hover:bg-[var(--secondary-hover)] hover:-translate-x-0.5 hover:-translate-y-0.5 transition-all duration-150 btn-press inline-flex items-center gap-2"
                >
                  Get the Bot ★
                </a>
              </div>

              <div className="border-t-2 border-dashed border-[var(--border-light)]" />

              {/* CTA */}
              <div className="pixel-border-double bg-[var(--muted)] scanlines p-4 text-center shadow-[4px_4px_0px_0px_var(--violet)]">
                <div className="flex items-center justify-center gap-2 mb-2" aria-hidden="true">
                  <Sparkle color="var(--violet)" />
                  <Sparkle color="var(--accent-bright)" />
                  <Sparkle color="var(--violet)" />
                </div>
                <p className="pixel-font text-[10px] text-[var(--foreground)] tracking-wide mb-3 leading-relaxed">
                  Submit your own translations!
                </p>
                <a
                  href={REPO_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="pixel-font text-[10px] px-3 py-1.5 bg-[var(--violet)] text-white border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm tracking-wide uppercase hover:bg-[var(--violet-deep)] hover:-translate-x-0.5 hover:-translate-y-0.5 transition-all duration-150 btn-press inline-flex items-center gap-2"
                >
                  <Github size={12} strokeWidth={2.5} />
                  GitHub
                </a>
              </div>

              {/* Spacer */}
              <div className="flex-1" />

              {/* Footer in sidebar */}
              <div className="text-center">
                <p className="pixel-font text-[10px] text-[var(--foreground-muted)] tracking-wide">
                  Made with <span className="text-[var(--primary)]">&#9829;</span> and pixels
                </p>
                <p className="mono-font text-[10px] text-[var(--border)] tracking-widest uppercase mt-1">
                  leagueofren.com
                </p>
              </div>
            </aside>

            {/* ── Content Area ── */}
            <main className="flex-1 p-4 lg:p-8">
              <Routes>
                <Route path="/" element={<Leaderboard />} />
              </Routes>
            </main>
          </div>

          {/* ── Mobile-only About + Footer ── */}
          <div className="lg:hidden">
            <div className="pixel-divider mx-6" aria-hidden="true" />

            <section className="bg-[var(--background-alt)] dither-check py-10">
              <div className="mx-auto max-w-lg px-6 space-y-4">
                <div className="pixel-border bg-[var(--card)] pixel-shadow-md p-5">
                  <div className="flex items-center gap-2 mb-3">
                    <div className="w-7 h-7 bg-[var(--sky)] border-4 border-[var(--border)] rounded-[4px] flex items-center justify-center pixel-shadow-sm">
                      <BookOpen size={14} strokeWidth={2.5} className="text-white" />
                    </div>
                    <h3 className="pixel-font text-xs tracking-wide">What is this?</h3>
                  </div>
                  <p className="text-sm text-[var(--foreground-muted)] leading-relaxed">
                    A Discord bot that translates Korean and Chinese summoner names in your League games.
                    This leaderboard collects the best translations from the community.
                  </p>
                </div>
                <div className="pixel-border bg-[var(--card)] pixel-shadow-md p-5">
                  <div className="flex items-center gap-2 mb-3">
                    <div className="w-7 h-7 bg-[var(--secondary)] border-4 border-[var(--border)] rounded-[4px] flex items-center justify-center pixel-shadow-sm">
                      <Heart size={14} strokeWidth={2.5} className="text-white" />
                    </div>
                    <h3 className="pixel-font text-xs tracking-wide">Contribute</h3>
                  </div>
                  <p className="text-sm text-[var(--foreground-muted)] leading-relaxed mb-3">
                    Run the bot and opt in to share translations.
                  </p>
                  <a
                    href={REPO_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="pixel-font text-[10px] px-3 py-1.5 bg-[var(--secondary)] text-white border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm tracking-wide uppercase btn-press inline-flex items-center gap-2"
                  >
                    Get the Bot ★
                  </a>
                </div>
              </div>
            </section>

            <footer className="bg-[var(--card)] dither-check py-6 border-t-4 border-[var(--border)]">
              <div className="px-6 flex items-center justify-between">
                <a href={REPO_URL} target="_blank" rel="noopener noreferrer" aria-label="GitHub"
                  className="w-10 h-10 bg-[var(--card)] border-4 border-[var(--violet)] rounded-[8px] shadow-[2px_2px_0px_0px_var(--violet)] flex items-center justify-center">
                  <Github size={18} strokeWidth={2.5} className="text-[var(--foreground)]" />
                </a>
                <p className="pixel-font text-xs text-[var(--foreground-muted)] tracking-wide">
                  Made with <span className="text-[var(--primary)]">&#9829;</span> and pixels
                </p>
              </div>
            </footer>
          </div>

        </div>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
