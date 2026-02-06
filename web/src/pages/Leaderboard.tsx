import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ChevronUp, ChevronDown } from 'lucide-react'
import { listTranslations, vote } from '../lib/api'
import type { SortOption, PeriodOption } from '../lib/schemas'

const SORT_OPTIONS: { value: SortOption; label: string }[] = [
  { value: 'hot', label: '🔥 Hot' },
  { value: 'new', label: '✨ New' },
  { value: 'top', label: '⭐ Top' },
]

const PERIOD_OPTIONS: { value: PeriodOption; label: string }[] = [
  { value: 'hour', label: 'Hour' },
  { value: 'day', label: 'Day' },
  { value: 'week', label: 'Week' },
  { value: 'month', label: 'Month' },
  { value: 'year', label: 'Year' },
  { value: 'all', label: 'All' },
]

const REGIONS = ['', 'NA', 'EUW', 'EUNE', 'KR', 'JP', 'BR', 'LAN', 'LAS', 'OCE', 'TR', 'RU', 'TW']
const LANGUAGES = ['', 'korean', 'chinese']

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

  return (
    <div className="space-y-6">
      {/* Sort Tabs */}
      <div className="flex gap-2">
        {SORT_OPTIONS.map(opt => (
          <button
            key={opt.value}
            onClick={() => { setSort(opt.value); setPage(1) }}
            className={`pixel-font text-sm px-4 py-2 pixel-border transition-all duration-150 ${
              sort === opt.value
                ? 'bg-[var(--primary)] text-white pixel-shadow-sm'
                : 'bg-[var(--card)] hover:bg-[var(--muted)] pixel-shadow-sm hover:-translate-y-0.5'
            }`}
          >
            {opt.label}
          </button>
        ))}
      </div>

      {/* Period selector (only for "top") */}
      {sort === 'top' && (
        <div className="flex gap-1 flex-wrap">
          {PERIOD_OPTIONS.map(opt => (
            <button
              key={opt.value}
              onClick={() => { setPeriod(opt.value); setPage(1) }}
              className={`text-xs px-3 py-1 border-2 border-[var(--border)] rounded-[4px] transition-colors ${
                period === opt.value
                  ? 'bg-[var(--accent)] font-bold'
                  : 'bg-[var(--card)] hover:bg-[var(--muted)]'
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>
      )}

      {/* Filters */}
      <div className="flex gap-3 flex-wrap">
        <select
          value={region}
          onChange={e => { setRegion(e.target.value); setPage(1) }}
          className="pixel-border bg-[var(--card)] px-3 py-2 text-sm"
        >
          <option value="">All Regions</option>
          {REGIONS.filter(Boolean).map(r => <option key={r} value={r}>{r}</option>)}
        </select>
        <select
          value={language}
          onChange={e => { setLanguage(e.target.value); setPage(1) }}
          className="pixel-border bg-[var(--card)] px-3 py-2 text-sm"
        >
          <option value="">All Languages</option>
          {LANGUAGES.filter(Boolean).map(l => <option key={l} value={l}>{l.charAt(0).toUpperCase() + l.slice(1)}</option>)}
        </select>
      </div>

      {/* Translation Cards */}
      {isLoading ? (
        <div className="pixel-font text-center text-[var(--foreground-muted)] py-12">Loading...</div>
      ) : (
        <div className="space-y-4">
          {data?.data.map(t => (
            <div key={t.id} className="pixel-border bg-[var(--card)] pixel-shadow-md p-4 flex items-center gap-4 hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[6px_6px_0px_0px_var(--border)] transition-all duration-200">
              {/* Vote buttons */}
              <div className="flex flex-col items-center gap-1 min-w-[48px]">
                <button
                  onClick={() => voteMutation.mutate({ id: t.id, direction: 1 })}
                  className="p-1 hover:text-[var(--secondary)] transition-colors"
                >
                  <ChevronUp size={24} strokeWidth={3} />
                </button>
                <span className="pixel-font text-sm font-bold">
                  {t.upvotes - t.downvotes}
                </span>
                <button
                  onClick={() => voteMutation.mutate({ id: t.id, direction: -1 })}
                  className="p-1 hover:text-[var(--destructive)] transition-colors"
                >
                  <ChevronDown size={24} strokeWidth={3} />
                </button>
              </div>

              {/* Translation content */}
              <div className="flex-1 min-w-0">
                <div className="flex items-baseline gap-3 flex-wrap">
                  <span className="pixel-font text-lg text-[var(--primary)]">{t.username}</span>
                  <span className="text-[var(--foreground-muted)]">&rarr;</span>
                  <span className="font-bold text-lg">{t.translation}</span>
                </div>
                {t.explanation && (
                  <p className="text-sm text-[var(--foreground-muted)] mt-1">{t.explanation}</p>
                )}
                <div className="flex gap-3 mt-2 text-xs text-[var(--foreground-muted)]">
                  <span className="px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px]">{t.region}</span>
                  <span className="px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px]">{t.language}</span>
                  {t.riot_verified && <span className="px-2 py-0.5 bg-[var(--secondary)] text-white border-2 border-[var(--border)] rounded-[4px]">&#10003; Verified</span>}
                </div>
              </div>
            </div>
          ))}

          {data?.data.length === 0 && (
            <div className="pixel-border bg-[var(--card)] pixel-shadow-md p-8 text-center">
              <p className="pixel-font text-[var(--foreground-muted)]">No translations yet!</p>
            </div>
          )}
        </div>
      )}

      {/* Pagination */}
      {data && data.pagination.total > 25 && (
        <div className="flex justify-center gap-2">
          <button
            disabled={page <= 1}
            onClick={() => setPage(p => p - 1)}
            className="pixel-border bg-[var(--card)] px-4 py-2 pixel-shadow-sm disabled:opacity-50 hover:-translate-y-0.5 transition-all"
          >
            &larr; Prev
          </button>
          <span className="pixel-font text-sm self-center">
            Page {page} of {Math.ceil(data.pagination.total / 25)}
          </span>
          <button
            disabled={page >= Math.ceil(data.pagination.total / 25)}
            onClick={() => setPage(p => p + 1)}
            className="pixel-border bg-[var(--card)] px-4 py-2 pixel-shadow-sm disabled:opacity-50 hover:-translate-y-0.5 transition-all"
          >
            Next &rarr;
          </button>
        </div>
      )}
    </div>
  )
}
