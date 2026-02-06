import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Github, Heart, BookOpen, Swords } from 'lucide-react'
import { Leaderboard } from './pages/Leaderboard'

const queryClient = new QueryClient()

const REPO_URL = 'https://github.com/jusunglee/leagueofren'

function Sparkle({ color = 'var(--accent-bright)' }: { color?: string }) {
  return <span className="pixel-sparkle" style={{ background: color }} aria-hidden="true" />
}

function SectionMarker({ label }: { label: string }) {
  return (
    <div className="flex items-center gap-3 mb-6">
      <div className="w-6 h-6 bg-[var(--violet)] border-4 border-[var(--border)] rounded-[4px] flex items-center justify-center pixel-shadow-sm">
        <span className="text-[8px] text-white font-bold" aria-hidden="true">&#9873;</span>
      </div>
      <span className="pixel-font text-xs tracking-widest uppercase text-[var(--foreground-muted)]">{label}</span>
      <div className="flex-1 border-t-2 border-dashed border-[var(--border-light)]" />
    </div>
  )
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="min-h-screen flex flex-col">

          {/* ── Top Banner ── */}
          <header className="pixel-border bg-[var(--card)] pixel-shadow-lg mx-6 lg:mx-auto lg:max-w-6xl mt-6 p-4 lg:p-5 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-[var(--violet)] border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm flex items-center justify-center">
                <Swords size={20} strokeWidth={2.5} className="text-white" />
              </div>
              <div>
                <h1 className="pixel-font text-base lg:text-xl text-[var(--violet)] tracking-wide leading-tight">
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
          </header>

          {/* ── Main Content ── */}
          <main className="mx-6 lg:mx-auto lg:max-w-6xl py-8 flex-1">
            <Routes>
              <Route path="/" element={<Leaderboard />} />
            </Routes>
          </main>

          <div className="pixel-divider mx-6 lg:mx-auto lg:max-w-6xl" aria-hidden="true" />

          {/* ── About + CTA (side by side on desktop) ── */}
          <section className="bg-[var(--background-alt)] dither-check py-12 lg:py-16">
            <div className="mx-auto max-w-6xl px-6">
              <SectionMarker label="About" />
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <div className="pixel-border bg-[var(--card)] pixel-shadow-md p-6">
                  <div className="flex items-center gap-2 mb-4">
                    <div className="w-8 h-8 bg-[var(--sky)] border-4 border-[var(--border)] rounded-[4px] flex items-center justify-center pixel-shadow-sm">
                      <BookOpen size={16} strokeWidth={2.5} className="text-white" />
                    </div>
                    <h3 className="pixel-font text-xs tracking-wide">What is this?</h3>
                  </div>
                  <p className="text-sm text-[var(--foreground-muted)] leading-relaxed">
                    A Discord bot that translates Korean and Chinese summoner names in your League games.
                    This leaderboard collects the best translations from the community.
                  </p>
                </div>
                <div className="pixel-border bg-[var(--card)] pixel-shadow-md p-6">
                  <div className="flex items-center gap-2 mb-4">
                    <div className="w-8 h-8 bg-[var(--secondary)] border-4 border-[var(--border)] rounded-[4px] flex items-center justify-center pixel-shadow-sm">
                      <Heart size={16} strokeWidth={2.5} className="text-white" />
                    </div>
                    <h3 className="pixel-font text-xs tracking-wide">Contribute</h3>
                  </div>
                  <p className="text-sm text-[var(--foreground-muted)] leading-relaxed mb-4">
                    Run the bot and opt in to share translations. Every submission helps build a richer
                    database for all players.
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
                {/* CTA card */}
                <div className="pixel-border-double bg-[var(--muted)] scanlines p-6 flex flex-col justify-center items-center text-center shadow-[4px_4px_0px_0px_var(--violet)]">
                  <div className="flex items-center gap-2 mb-3" aria-hidden="true">
                    <Sparkle color="var(--violet)" />
                    <Sparkle color="var(--accent-bright)" />
                    <Sparkle color="var(--violet)" />
                  </div>
                  <p className="pixel-font text-xs text-[var(--foreground)] tracking-wide mb-4 leading-relaxed">
                    Submit your own translations!
                  </p>
                  <a
                    href={REPO_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="pixel-font text-[10px] px-4 py-2 bg-[var(--violet)] text-white border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm tracking-wide uppercase hover:bg-[var(--violet-deep)] hover:-translate-x-0.5 hover:-translate-y-0.5 transition-all duration-150 btn-press inline-flex items-center gap-2"
                  >
                    <Github size={14} strokeWidth={2.5} />
                    GitHub
                  </a>
                </div>
              </div>
            </div>
          </section>

          {/* ── Footer ── */}
          <footer className="bg-[var(--card)] dither-check py-6 border-t-4 border-[var(--border)]">
            <div className="mx-auto max-w-6xl px-6">
              <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
                <a
                  href={REPO_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="w-10 h-10 bg-[var(--card)] border-4 border-[var(--violet)] rounded-[8px] shadow-[2px_2px_0px_0px_var(--violet)] flex items-center justify-center hover:-translate-y-0.5 transition-all duration-150"
                  aria-label="GitHub"
                >
                  <Github size={18} strokeWidth={2.5} className="text-[var(--foreground)]" />
                </a>
                <p className="pixel-font text-xs text-[var(--foreground-muted)] tracking-wide">
                  Made with <span className="text-[var(--primary)]">&#9829;</span> and pixels
                </p>
                <p className="mono-font text-xs text-[var(--border)] tracking-widest uppercase">
                  leagueofren.com
                </p>
              </div>
            </div>
          </footer>

        </div>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
