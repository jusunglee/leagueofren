import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ChevronUp, ChevronDown, Flame, Sparkles, Trophy } from 'lucide-react'
import { listTranslations, vote } from '../lib/api'
import type { SortOption, PeriodOption, Translation } from '../lib/schemas'

const SORT_OPTIONS: { value: SortOption; label: string; icon: typeof Flame }[] = [
  { value: 'hot', label: 'Hot', icon: Flame },
  { value: 'new', label: 'New', icon: Sparkles },
  { value: 'top', label: 'Top', icon: Trophy },
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
          <span className="mono-font text-xs px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase">
            {t.region}
          </span>
          <span className="mono-font text-xs px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase">
            {t.language}
          </span>
          {t.riot_verified && (
            <span className="pixel-font text-[10px] px-2 py-0.5 bg-[var(--secondary)] text-white border-2 border-[var(--border)] rounded-[4px] tracking-wide">
              &#10003; Verified
            </span>
          )}
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

      {/* Controls row — all inline on desktop */}
      <div className="flex flex-wrap items-center gap-3">
        {/* Sort Tabs */}
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
              <Icon size={14} strokeWidth={2.5} />
              {opt.label}
            </button>
          )
        })}

        {/* Period (inline, shown when top) */}
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

        {/* Spacer */}
        <div className="flex-1" />

        {/* Filters (right-aligned on desktop) */}
        <select
          value={region}
          onChange={e => { setRegion(e.target.value); setPage(1) }}
          className="pixel-border bg-[var(--card)] px-3 py-1.5 text-sm focus:border-[var(--violet)] focus:outline-none"
        >
          <option value="">All Regions</option>
          {REGIONS.map(r => <option key={r} value={r}>{r}</option>)}
        </select>
        <select
          value={language}
          onChange={e => { setLanguage(e.target.value); setPage(1) }}
          className="pixel-border bg-[var(--card)] px-3 py-1.5 text-sm focus:border-[var(--violet)] focus:outline-none"
        >
          <option value="">All Languages</option>
          {LANGUAGES.map(l => <option key={l} value={l}>{l.charAt(0).toUpperCase() + l.slice(1)}</option>)}
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
