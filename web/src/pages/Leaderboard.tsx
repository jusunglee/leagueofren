import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ChevronUp, ChevronDown, Flame, Sparkles, Trophy, ExternalLink, MessageCircleQuestion } from 'lucide-react'
import { listTranslations, vote } from '../lib/api'
import type { SortOption, PeriodOption, Translation } from '../lib/schemas'

const SORT_OPTIONS: { value: SortOption; label: string; icon: typeof Flame; color: string }[] = [
  { value: 'hot', label: 'Hot', icon: Flame, color: '#E85D75' },
  { value: 'new', label: 'New', icon: Sparkles, color: '#FFD93D' },
  { value: 'top', label: 'Top', icon: Trophy, color: '#F2A65A' },
]

const PERIOD_OPTIONS: { value: PeriodOption; label: string }[] = [
  { value: 'hour', label: 'Hour' },
  { value: 'day', label: 'Day' },
  { value: 'week', label: 'Week' },
  { value: 'month', label: 'Month' },
  { value: 'year', label: 'Year' },
  { value: 'all', label: 'All Time' },
]

const REGIONS = ['NA', 'EUW', 'EUNE', 'KR', 'JP', 'BR', 'LAN', 'LAS', 'OCE', 'TR', 'RU', 'TW']
const LANGUAGES = ['korean', 'chinese']

const BADGE_COLORS = [
  'var(--primary)', 'var(--violet)', 'var(--accent)',
  'var(--sky)', 'var(--lavender)', 'var(--secondary)',
]

const REGION_EMOJI: Record<string, string> = {
  NA: '🇺🇸', EUW: '🇪🇺', EUNE: '🇪🇺', KR: '🇰🇷', JP: '🇯🇵',
  BR: '🇧🇷', LAN: '🌎', LAS: '🌎', OCE: '🇦🇺', TR: '🇹🇷', RU: '🇷🇺', TW: '🇹🇼',
}

const LANGUAGE_EMOJI: Record<string, string> = {
  korean: '🇰🇷',
  chinese: '🇨🇳',
}

function RankBadge({ index }: { index: number }) {
  const color = BADGE_COLORS[index % BADGE_COLORS.length]
  return (
    <div
      className="absolute -top-2 -right-2 w-6 h-6 border-4 border-[var(--border)] rounded-[4px] flex items-center justify-center pixel-shadow-sm"
      style={{ background: color }}
      aria-hidden="true"
    >
      <span className="pixel-font text-[8px] text-white leading-none">{index + 1}</span>
    </div>
  )
}

function buildLearnMoreUrl(username: string, translation: string, explanation: string | null) {
  const prompt = `Tell me more about the League of Legends summoner name "${username}". It translates to "${translation}"${explanation ? ` (${explanation})` : ''}. What's the cultural context, any references to games, anime, literature, or memes? Is this a common naming pattern?`
  return `https://chatgpt.com/?q=${encodeURIComponent(prompt)}`
}

function buildOpggUrl(username: string, region: string) {
  const regionSlug = region.toLowerCase()
  return `https://www.op.gg/summoners/${regionSlug}/${encodeURIComponent(username)}`
}

function buildPorofessorUrl(username: string, region: string) {
  const regionSlug = region.toLowerCase()
  return `https://www.porofessor.gg/live/${regionSlug}/${encodeURIComponent(username)}`
}

function TranslationCard({ t, index, onVote }: {
  t: Translation
  index: number
  onVote: (id: number, dir: 1 | -1) => void
}) {
  const score = t.upvotes - t.downvotes
  const isTop3 = index < 3

  return (
    <div
      className={`relative pixel-border bg-[var(--card)] p-4 lg:p-5 flex items-center gap-4 transition-all duration-200 animate-fade-in ${
        isTop3
          ? 'pixel-border-double bg-[var(--background-alt)] shadow-[6px_6px_0px_0px_var(--violet)] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[8px_8px_0px_0px_var(--violet)]'
          : 'pixel-shadow-md hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[6px_6px_0px_0px_var(--border)]'
      }`}
      style={{ animationDelay: `${index * 40}ms` }}
    >
      <RankBadge index={index} />

      {/* Vote column */}
      <div className="flex flex-col items-center gap-0 min-w-[48px]">
        <button
          onClick={() => onVote(t.id, 1)}
          className="w-10 h-10 flex items-center justify-center border-2 border-transparent rounded-[4px] hover:border-[var(--secondary)] hover:bg-[var(--muted)] transition-all duration-150 focus-visible:ring-2 focus-visible:ring-[var(--ring)]"
          aria-label="Upvote"
        >
          <ChevronUp size={24} strokeWidth={3} />
        </button>
        <span className={`pixel-font text-sm leading-none py-1 ${
          score > 0 ? 'text-[var(--secondary)]' : score < 0 ? 'text-[var(--destructive)]' : 'text-[var(--foreground-muted)]'
        }`}>
          {score}
        </span>
        <button
          onClick={() => onVote(t.id, -1)}
          className="w-10 h-10 flex items-center justify-center border-2 border-transparent rounded-[4px] hover:border-[var(--destructive)] hover:bg-[var(--muted)] transition-all duration-150 focus-visible:ring-2 focus-visible:ring-[var(--ring)]"
          aria-label="Downvote"
        >
          <ChevronDown size={24} strokeWidth={3} />
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <div className="flex items-baseline gap-3 flex-wrap">
          <span className="pixel-font text-base lg:text-lg text-[var(--primary)] tracking-wide">{t.username}</span>
          <span className="text-[var(--border)] font-bold">&rarr;</span>
          <span className="font-bold text-base lg:text-lg">{t.translation}</span>
        </div>
        {t.explanation && (
          <p className="text-sm text-[var(--foreground-muted)] mt-1 leading-relaxed">{t.explanation}</p>
        )}
        <div className="flex items-center gap-2 mt-3 flex-wrap">
          {/* Region badge with flag */}
          <span className="mono-font text-xs px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1">
            <span className="text-sm leading-none" aria-hidden="true">{REGION_EMOJI[t.region] || '🌍'}</span>
            {t.region}
          </span>
          {/* Language badge with flag */}
          <span className="mono-font text-xs px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1">
            <span className="text-sm leading-none" aria-hidden="true">{LANGUAGE_EMOJI[t.language] || '🌐'}</span>
            {t.language}
          </span>

          {/* Spacer for link badges */}
          <span className="hidden lg:inline w-px h-4 bg-[var(--border-light)]" />

          {/* OP.GG badge */}
          <a
            href={buildOpggUrl(t.username, t.region)}
            target="_blank"
            rel="noopener noreferrer"
            className="mono-font text-[10px] px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1 hover:bg-[var(--violet)] hover:text-white hover:border-[var(--border)] transition-all duration-150"
          >
            <ExternalLink size={10} strokeWidth={2.5} />
            OP.GG
          </a>
          {/* Porofessor badge */}
          <a
            href={buildPorofessorUrl(t.username, t.region)}
            target="_blank"
            rel="noopener noreferrer"
            className="mono-font text-[10px] px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1 hover:bg-[var(--sky)] hover:text-white hover:border-[var(--border)] transition-all duration-150"
          >
            <ExternalLink size={10} strokeWidth={2.5} />
            Porofessor
          </a>
          {/* Learn More (ChatGPT) */}
          <a
            href={buildLearnMoreUrl(t.username, t.translation, t.explanation)}
            target="_blank"
            rel="noopener noreferrer"
            className="mono-font text-[10px] px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1 hover:bg-[var(--accent)] hover:text-[var(--background)] hover:border-[var(--border)] transition-all duration-150"
          >
            <MessageCircleQuestion size={10} strokeWidth={2.5} />
            Learn More
          </a>
        </div>
      </div>
    </div>
  )
}

function SectionMarker({ label }: { label: string }) {
  return (
    <div className="flex items-center gap-3 mb-6">
      <div className="w-6 h-6 bg-[var(--primary)] border-4 border-[var(--border)] rounded-[4px] flex items-center justify-center pixel-shadow-sm">
        <span className="text-[8px] text-white font-bold" aria-hidden="true">&#9873;</span>
      </div>
      <span className="pixel-font text-xs tracking-widest uppercase text-[var(--foreground-muted)]">{label}</span>
      <div className="flex-1 border-t-2 border-dashed border-[var(--border-light)]" />
    </div>
  )
}

export function Leaderboard() {
  const queryClient = useQueryClient()
  const [sort, setSort] = useState<SortOption>('hot')
  const [period, setPeriod] = useState<PeriodOption>('week')
  const [region, setRegion] = useState('')
  const [language, setLanguage] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['translations', sort, period, region, language, page],
    queryFn: () => listTranslations({ sort, period, region, language, page, limit: 25 }),
  })

  const voteMutation = useMutation({
    mutationFn: ({ id, direction }: { id: number; direction: 1 | -1 }) => vote(id, direction),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['translations'] }),
  })

  const totalPages = data ? Math.ceil(data.pagination.total / 25) : 0

  return (
    <div className="space-y-6">
      <SectionMarker label="Rankings" />

      {/* Controls row */}
      <div className="flex flex-wrap items-center gap-3">
        {SORT_OPTIONS.map(opt => {
          const Icon = opt.icon
          return (
            <button
              key={opt.value}
              onClick={() => { setSort(opt.value); setPage(1) }}
              className={`pixel-font text-xs px-4 lg:px-5 py-2 pixel-border transition-all duration-150 btn-press inline-flex items-center gap-2 focus-visible:ring-2 focus-visible:ring-[var(--ring)] focus-visible:ring-offset-2 ${
                sort === opt.value
                  ? 'bg-[var(--violet)] text-white pixel-shadow-sm'
                  : 'bg-[var(--card)] pixel-shadow-sm hover:bg-[var(--muted)] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_0px_var(--border)]'
              }`}
            >
              <Icon size={14} strokeWidth={2.5} style={{ color: sort === opt.value ? 'white' : opt.color }} />
              {opt.label}
            </button>
          )
        })}

        {sort === 'top' && (
          <>
            <div className="w-px h-6 bg-[var(--border-light)]" />
            {PERIOD_OPTIONS.map(opt => (
              <button
                key={opt.value}
                onClick={() => { setPeriod(opt.value); setPage(1) }}
                className={`text-xs px-3 py-1.5 border-2 border-[var(--border)] rounded-[4px] transition-all duration-150 focus-visible:ring-2 focus-visible:ring-[var(--ring)] ${
                  period === opt.value
                    ? 'bg-[var(--accent)] text-[var(--background)] font-bold pixel-shadow-sm'
                    : 'bg-[var(--card)] hover:bg-[var(--muted)]'
                }`}
              >
                {opt.label}
              </button>
            ))}
          </>
        )}

        <div className="flex-1" />

        {/* Filters with more spacing */}
        <div className="flex items-center gap-4">
          <select
            value={region}
            onChange={e => { setRegion(e.target.value); setPage(1) }}
            className="pixel-border bg-[var(--card)] px-4 py-1.5 text-sm focus:border-[var(--violet)] focus:outline-none"
          >
            <option value="">All Regions</option>
            {REGIONS.map(r => <option key={r} value={r}>{REGION_EMOJI[r] || ''} {r}</option>)}
          </select>
          <select
            value={language}
            onChange={e => { setLanguage(e.target.value); setPage(1) }}
            className="pixel-border bg-[var(--card)] px-4 py-1.5 text-sm focus:border-[var(--violet)] focus:outline-none"
          >
            <option value="">All Languages</option>
            {LANGUAGES.map(l => <option key={l} value={l}>{LANGUAGE_EMOJI[l] || ''} {l.charAt(0).toUpperCase() + l.slice(1)}</option>)}
          </select>
          {(region || language) && (
            <button
              onClick={() => { setRegion(''); setLanguage(''); setPage(1) }}
              className="text-xs px-3 py-1.5 border-2 border-dashed border-[var(--border-light)] rounded-[4px] text-[var(--foreground-muted)] hover:border-solid hover:bg-[var(--muted)] transition-all"
            >
              Clear
            </button>
          )}
        </div>
      </div>

      {/* Translation Cards */}
      {isLoading ? (
        <div className="pixel-border bg-[var(--card)] pixel-shadow-md p-12 text-center">
          <p className="pixel-font text-sm text-[var(--foreground-muted)] tracking-wide">Loading translations...</p>
        </div>
      ) : (
        <div className="space-y-3">
          {data?.data.map((t, i) => (
            <TranslationCard
              key={t.id}
              t={t}
              index={(page - 1) * 25 + i}
              onVote={(id, dir) => voteMutation.mutate({ id, direction: dir })}
            />
          ))}

          {data?.data.length === 0 && (
            <div className="pixel-border-double bg-[var(--background-alt)] pixel-shadow-violet p-8 lg:p-12 text-center">
              <p className="pixel-font text-sm text-[var(--foreground-muted)] mb-2">No translations found!</p>
              <p className="text-sm text-[var(--foreground-muted)]">Try adjusting your filters or check back later.</p>
            </div>
          )}
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-4">
          <button
            disabled={page <= 1}
            onClick={() => setPage(p => p - 1)}
            className="pixel-font text-xs px-4 py-2 pixel-border bg-[var(--card)] pixel-shadow-sm disabled:opacity-40 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_0px_var(--border)] transition-all duration-150 btn-press tracking-wide focus-visible:ring-2 focus-visible:ring-[var(--ring)]"
          >
            &larr; Prev
          </button>
          <span className="mono-font text-xs tracking-widest text-[var(--foreground-muted)]">
            {page} / {totalPages}
          </span>
          <button
            disabled={page >= totalPages}
            onClick={() => setPage(p => p + 1)}
            className="pixel-font text-xs px-4 py-2 pixel-border bg-[var(--card)] pixel-shadow-sm disabled:opacity-40 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_0px_var(--border)] transition-all duration-150 btn-press tracking-wide focus-visible:ring-2 focus-visible:ring-[var(--ring)]"
          >
            Next &rarr;
          </button>
        </div>
      )}
    </div>
  )
}
