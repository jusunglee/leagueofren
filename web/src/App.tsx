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
      <div className="w-6 h-6 bg-[var(--accent)] border-4 border-[var(--border)] rounded-[4px] flex items-center justify-center pixel-shadow-sm">
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

          {/* ── Header ── */}
          <header className="pixel-border bg-[var(--card)] pixel-shadow-lg mx-4 mt-4 md:mx-auto md:max-w-5xl p-4 md:p-6 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-[var(--primary)] border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm flex items-center justify-center">
                <Swords size={20} strokeWidth={2.5} className="text-white" />
              </div>
              <div>
                <h1 className="pixel-font text-base md:text-2xl text-[var(--primary)] tracking-wide leading-tight">
                  League of Ren
                </h1>
                <p className="text-xs text-[var(--foreground-muted)] hidden sm:block">Translation Rankings</p>
              </div>
            </div>
            <a
              href={REPO_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="w-10 h-10 bg-[var(--card)] border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm flex items-center justify-center hover:bg-[var(--muted)] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_0px_var(--border)] transition-all duration-150 btn-press"
              aria-label="GitHub Repository"
            >
              <Github size={20} strokeWidth={2.5} />
            </a>
          </header>

          {/* ── Hero ── */}
          <section className="mx-4 md:mx-auto md:max-w-5xl mt-8 mb-4 text-center py-8 md:py-12">
            <div className="flex items-center justify-center gap-2 mb-4">
              <Sparkle color="var(--primary)" />
              <Sparkle color="var(--accent)" />
              <Sparkle color="var(--sky)" />
            </div>
            <h2 className="pixel-font text-2xl md:text-4xl lg:text-5xl text-[var(--foreground)] tracking-wide leading-tight mb-4">
              What do those names{' '}
              <span className="text-[var(--primary)]">actually</span>{' '}
              mean?
            </h2>
            <p className="text-lg md:text-xl text-[var(--foreground-muted)] max-w-2xl mx-auto leading-relaxed mb-8">
              Community-ranked translations of Korean and Chinese League of Legends summoner names.
              Vote for the best translations, discover hidden meanings, and learn what your opponents' names really say.
            </p>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <a
                href={REPO_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="w-full sm:w-auto pixel-font text-sm px-6 py-3 bg-[var(--primary)] text-white border-4 border-[var(--border)] rounded-[8px] pixel-shadow-md tracking-wide uppercase hover:bg-[var(--primary-hover)] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[6px_6px_0px_0px_var(--border)] transition-all duration-150 btn-press inline-flex items-center justify-center gap-2 focus-visible:ring-2 focus-visible:ring-[var(--ring)] focus-visible:ring-offset-2 focus-visible:ring-offset-[var(--background)]"
              >
                Submit Your Own ★
              </a>
              <span className="hidden md:inline-block animate-[bounce-cursor_1.5s_ease-in-out_infinite] text-[var(--accent)]" aria-hidden="true">&#9654;</span>
            </div>
          </section>

          <div className="pixel-divider mx-4 md:mx-auto md:max-w-5xl" aria-hidden="true" />

          {/* ── Main Content ── */}
          <main className="mx-4 md:mx-auto md:max-w-5xl py-8 flex-1">
            <Routes>
              <Route path="/" element={<Leaderboard />} />
            </Routes>
          </main>

          <div className="pixel-divider mx-4 md:mx-auto md:max-w-5xl" aria-hidden="true" />

          {/* ── About Section (dithered background) ── */}
          <section className="bg-[var(--background-alt)] dither-check py-12 md:py-16">
            <div className="mx-auto max-w-5xl px-6">
              <SectionMarker label="About" />
              <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                <div className="pixel-border bg-[var(--card)] pixel-shadow-md p-6">
                  <div className="flex items-center gap-2 mb-4">
                    <div className="w-8 h-8 bg-[var(--sky)] border-4 border-[var(--border)] rounded-[4px] flex items-center justify-center pixel-shadow-sm">
                      <BookOpen size={16} strokeWidth={2.5} className="text-white" />
                    </div>
                    <h3 className="pixel-font text-sm tracking-wide">What is this?</h3>
                  </div>
                  <p className="text-[var(--foreground-muted)] leading-relaxed">
                    League of Ren is a Discord bot that automatically translates Korean and Chinese
                    summoner names when your subscribed players enter a game. This leaderboard collects
                    the best translations from the community so everyone can discover the hidden
                    meanings behind the names they see in-game.
                  </p>
                </div>
                <div className="pixel-border bg-[var(--card)] pixel-shadow-md p-6">
                  <div className="flex items-center gap-2 mb-4">
                    <div className="w-8 h-8 bg-[var(--secondary)] border-4 border-[var(--border)] rounded-[4px] flex items-center justify-center pixel-shadow-sm">
                      <Heart size={16} strokeWidth={2.5} className="text-white" />
                    </div>
                    <h3 className="pixel-font text-sm tracking-wide">Want to contribute?</h3>
                  </div>
                  <p className="text-[var(--foreground-muted)] leading-relaxed mb-4">
                    Run the bot yourself and opt in to share your translations with the community.
                    Every translation you share helps build a richer database of name meanings for
                    all League players.
                  </p>
                  <a
                    href={REPO_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="pixel-font text-xs px-4 py-2 bg-[var(--secondary)] text-white border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm tracking-wide uppercase hover:bg-[var(--secondary-hover)] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_0px_var(--border)] transition-all duration-150 btn-press inline-flex items-center gap-2 focus-visible:ring-2 focus-visible:ring-[var(--ring)] focus-visible:ring-offset-2"
                  >
                    Get the Bot ★
                  </a>
                </div>
              </div>
            </div>
          </section>

          {/* ── CTA Banner (inverted/dark) ── */}
          <section className="bg-[var(--foreground)] scanlines border-y-4 border-[var(--accent)] py-12 md:py-16">
            <div className="mx-auto max-w-5xl px-6 text-center">
              <div className="flex items-center justify-center gap-2 mb-4" aria-hidden="true">
                <Sparkle color="var(--accent-bright)" />
                <Sparkle color="var(--accent)" />
                <Sparkle color="var(--accent-bright)" />
              </div>
              <h2 className="pixel-font text-xl md:text-3xl text-[var(--background)] tracking-wide leading-tight mb-4">
                Submit your own translations!
              </h2>
              <p className="text-[var(--border-light)] max-w-xl mx-auto leading-relaxed mb-8">
                Run the League of Ren Discord bot, opt in to sharing, and your translations
                will appear here for the community to vote on.
              </p>
              <a
                href={REPO_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="pixel-font text-sm px-8 py-4 bg-[var(--accent)] text-[var(--foreground)] border-4 border-[var(--accent-bright)] rounded-[8px] shadow-[4px_4px_0px_0px_var(--accent-bright)] tracking-wide uppercase hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[6px_6px_0px_0px_var(--accent-bright)] transition-all duration-150 btn-press inline-flex items-center gap-2 focus-visible:ring-2 focus-visible:ring-[var(--accent-bright)] focus-visible:ring-offset-2 focus-visible:ring-offset-[var(--foreground)]"
              >
                <Github size={16} strokeWidth={2.5} />
                View on GitHub ★
              </a>
            </div>
          </section>

          {/* ── Footer (cozy guestbook energy) ── */}
          <footer className="bg-[var(--foreground)] dither-check py-8 border-t-4 border-[var(--border)]">
            <div className="mx-auto max-w-5xl px-6">
              <div className="flex flex-col md:flex-row items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                  <a
                    href={REPO_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="w-10 h-10 bg-[var(--foreground)] border-4 border-[var(--accent)] rounded-[8px] shadow-[2px_2px_0px_0px_var(--accent)] flex items-center justify-center hover:-translate-y-0.5 transition-all duration-150"
                    aria-label="GitHub"
                  >
                    <Github size={18} strokeWidth={2.5} className="text-[var(--background)]" />
                  </a>
                </div>
                <p className="pixel-font text-xs text-[var(--border-light)] tracking-wide">
                  Made with <span className="text-[var(--primary)]">&#9829;</span> and pixels
                </p>
                <p className="mono-font text-xs text-[var(--foreground-muted)] tracking-widest uppercase">
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
