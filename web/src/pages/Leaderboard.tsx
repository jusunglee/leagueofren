import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ChevronUp, ChevronDown, MessageCircleQuestion } from 'lucide-react'
import { listTranslations, vote } from '../lib/api'
import type { SortOption, PeriodOption, Translation } from '../lib/schemas'

// Filled SVG icons for sort tabs
function FlameIcon({ filled, color }: { filled: boolean; color: string }) {
  return filled ? (
    <svg width="14" height="14" viewBox="0 0 24 24" fill={color} stroke={color} strokeWidth="2"><path d="M8.5 14.5A2.5 2.5 0 0 0 11 12c0-1.38-.5-2-1-3-1.072-2.143-.224-4.054 2-6 .5 2.5 2 4.9 4 6.5 2 1.6 3 3.5 3 5.5a7 7 0 1 1-14 0c0-1.153.433-2.294 1-3a2.5 2.5 0 0 0 2.5 2.5z"/></svg>
  ) : (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M8.5 14.5A2.5 2.5 0 0 0 11 12c0-1.38-.5-2-1-3-1.072-2.143-.224-4.054 2-6 .5 2.5 2 4.9 4 6.5 2 1.6 3 3.5 3 5.5a7 7 0 1 1-14 0c0-1.153.433-2.294 1-3a2.5 2.5 0 0 0 2.5 2.5z"/></svg>
  )
}

function SparklesIcon({ filled, color }: { filled: boolean; color: string }) {
  return filled ? (
    <svg width="14" height="14" viewBox="0 0 24 24" fill={color} stroke={color} strokeWidth="2"><path d="M9.937 15.5A2 2 0 0 0 8.5 14.063l-6.135-1.582a.5.5 0 0 1 0-.962L8.5 9.936A2 2 0 0 0 9.937 8.5l1.582-6.135a.5.5 0 0 1 .963 0L14.063 8.5A2 2 0 0 0 15.5 9.937l6.135 1.581a.5.5 0 0 1 0 .964L15.5 14.063a2 2 0 0 0-1.437 1.437l-1.582 6.135a.5.5 0 0 1-.963 0z"/><path d="M20 3v4"/><path d="M22 5h-4"/></svg>
  ) : (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M9.937 15.5A2 2 0 0 0 8.5 14.063l-6.135-1.582a.5.5 0 0 1 0-.962L8.5 9.936A2 2 0 0 0 9.937 8.5l1.582-6.135a.5.5 0 0 1 .963 0L14.063 8.5A2 2 0 0 0 15.5 9.937l6.135 1.581a.5.5 0 0 1 0 .964L15.5 14.063a2 2 0 0 0-1.437 1.437l-1.582 6.135a.5.5 0 0 1-.963 0z"/><path d="M20 3v4"/><path d="M22 5h-4"/></svg>
  )
}

function TrophyIcon({ filled, color }: { filled: boolean; color: string }) {
  return filled ? (
    <svg width="14" height="14" viewBox="0 0 24 24" fill={color} stroke={color} strokeWidth="2"><path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6"/><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18"/><path d="M4 22h16"/><path d="M10 14.66V17c0 .55-.47.98-.97 1.21C7.85 18.75 7 20.24 7 22"/><path d="M14 14.66V17c0 .55.47.98.97 1.21C16.15 18.75 17 20.24 17 22"/><path d="M18 2H6v7a6 6 0 0 0 12 0V2Z"/></svg>
  ) : (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6"/><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18"/><path d="M4 22h16"/><path d="M10 14.66V17c0 .55-.47.98-.97 1.21C7.85 18.75 7 20.24 7 22"/><path d="M14 14.66V17c0 .55.47.98.97 1.21C16.15 18.75 17 20.24 17 22"/><path d="M18 2H6v7a6 6 0 0 0 12 0V2Z"/></svg>
  )
}

// Tiny inline SVG logos for OP.GG and Porofessor
function OpggIcon() {
  return (
    <svg width="12" height="12" viewBox="0 0 24 24" aria-hidden="true">
      <rect width="24" height="24" rx="4" fill="#5383E8"/>
      <text x="12" y="17" textAnchor="middle" fill="white" fontSize="12" fontWeight="bold" fontFamily="sans-serif">O</text>
    </svg>
  )
}

function PorofessorIcon() {
  return (
    <svg width="12" height="12" viewBox="0 0 24 24" aria-hidden="true">
      <rect width="24" height="24" rx="4" fill="#785A28"/>
      <text x="12" y="17" textAnchor="middle" fill="white" fontSize="12" fontWeight="bold" fontFamily="sans-serif">P</text>
    </svg>
  )
}

const SORT_OPTIONS: { value: SortOption; label: string; color: string; icon: 'flame' | 'sparkles' | 'trophy' }[] = [
  { value: 'hot', label: 'Hot', color: '#E85D75', icon: 'flame' },
  { value: 'new', label: 'New', color: '#FFD93D', icon: 'sparkles' },
  { value: 'top', label: 'Top', color: '#F2A65A', icon: 'trophy' },
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

const RANK_COLORS: Record<string, string> = {
  IRON: '#5e5146',
  BRONZE: '#a0715e',
  SILVER: '#8c9ca8',
  GOLD: '#d4a634',
  PLATINUM: '#4e9e8e',
  EMERALD: '#3d9e5c',
  DIAMOND: '#576bce',
  MASTER: '#9d48c2',
  GRANDMASTER: '#d44545',
  CHALLENGER: '#f4c542',
}

const RANK_EMBLEM_URL = 'https://raw.communitydragon.org/latest/plugins/rcp-fe-lol-static-assets/global/default/ranked-emblem'

const RANK_ICON: Record<string, string> = {
  IRON: `${RANK_EMBLEM_URL}/emblem-iron.png`,
  BRONZE: `${RANK_EMBLEM_URL}/emblem-bronze.png`,
  SILVER: `${RANK_EMBLEM_URL}/emblem-silver.png`,
  GOLD: `${RANK_EMBLEM_URL}/emblem-gold.png`,
  PLATINUM: `${RANK_EMBLEM_URL}/emblem-platinum.png`,
  EMERALD: `${RANK_EMBLEM_URL}/emblem-emerald.png`,
  DIAMOND: `${RANK_EMBLEM_URL}/emblem-diamond.png`,
  MASTER: `${RANK_EMBLEM_URL}/emblem-master.png`,
  GRANDMASTER: `${RANK_EMBLEM_URL}/emblem-grandmaster.png`,
  CHALLENGER: `${RANK_EMBLEM_URL}/emblem-challenger.png`,
}

const ALL_RANKS = ['IRON', 'BRONZE', 'SILVER', 'GOLD', 'PLATINUM', 'EMERALD', 'DIAMOND', 'MASTER', 'GRANDMASTER', 'CHALLENGER']

function SortIcon({ type, filled, color }: { type: 'flame' | 'sparkles' | 'trophy'; filled: boolean; color: string }) {
  switch (type) {
    case 'flame': return <FlameIcon filled={filled} color={color} />
    case 'sparkles': return <SparklesIcon filled={filled} color={color} />
    case 'trophy': return <TrophyIcon filled={filled} color={color} />
  }
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
  const rank = t.rank?.toUpperCase()

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
          {/* Region badge */}
          <span className="mono-font text-xs px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1">
            <span className="text-sm leading-none" aria-hidden="true">{REGION_EMOJI[t.region] || '🌍'}</span>
            {t.region}
          </span>
          {/* Language badge */}
          <span className="mono-font text-xs px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1">
            <span className="text-sm leading-none" aria-hidden="true">{LANGUAGE_EMOJI[t.language] || '🌐'}</span>
            {t.language}
          </span>
          {/* Rank badge with official emblem */}
          {rank && RANK_COLORS[rank] && (
            <span
              className="mono-font text-[10px] px-2 py-0.5 border-2 rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1 font-bold text-white"
              style={{ background: RANK_COLORS[rank], borderColor: 'var(--border)' }}
            >
              {RANK_ICON[rank] && <img src={RANK_ICON[rank]} alt="" className="w-4 h-4 object-contain" aria-hidden="true" />}
              {rank}
            </span>
          )}

          <span className="hidden lg:inline w-px h-4 bg-[var(--border-light)]" />

          {/* OP.GG */}
          <a
            href={buildOpggUrl(t.username, t.region)}
            target="_blank"
            rel="noopener noreferrer"
            className="mono-font text-[10px] px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1 hover:bg-[#5383E8] hover:text-white hover:border-[var(--border)] transition-all duration-150"
          >
            <OpggIcon />
            OP.GG
          </a>
          {/* Porofessor */}
          <a
            href={buildPorofessorUrl(t.username, t.region)}
            target="_blank"
            rel="noopener noreferrer"
            className="mono-font text-[10px] px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1 hover:bg-[#785A28] hover:text-white hover:border-[var(--border)] transition-all duration-150"
          >
            <PorofessorIcon />
            Porofessor
          </a>
          {/* Learn More */}
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
  const [rank, setRank] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['translations', sort, period, region, language, page],
    queryFn: () => listTranslations({ sort, period, region, language, page, limit: 25 }),
  })

  const voteMutation = useMutation({
    mutationFn: ({ id, direction }: { id: number; direction: 1 | -1 }) => vote(id, direction),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['translations'] }),
  })

  const filteredData = rank && data?.data
    ? data.data.filter(t => t.rank?.toUpperCase() === rank)
    : data?.data

  const totalPages = data ? Math.ceil(data.pagination.total / 25) : 0

  return (
    <div className="space-y-6">
      <SectionMarker label="Rankings" />

      {/* Controls row */}
      <div className="flex flex-wrap items-center gap-3">
        {SORT_OPTIONS.map(opt => {
          const isActive = sort === opt.value
          return (
            <button
              key={opt.value}
              onClick={() => { setSort(opt.value); setPage(1) }}
              className={`pixel-font text-xs px-4 lg:px-5 py-2 pixel-border transition-all duration-150 btn-press inline-flex items-center gap-2 focus-visible:ring-2 focus-visible:ring-[var(--ring)] focus-visible:ring-offset-2 ${
                isActive
                  ? 'bg-[var(--violet)] text-white pixel-shadow-sm'
                  : 'bg-[var(--card)] pixel-shadow-sm hover:bg-[var(--muted)] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_0px_var(--border)]'
              }`}
            >
              <SortIcon type={opt.icon} filled={isActive} color={isActive ? 'white' : opt.color} />
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
          <select
            value={rank}
            onChange={e => { setRank(e.target.value); setPage(1) }}
            className="pixel-border bg-[var(--card)] px-4 py-1.5 text-sm focus:border-[var(--violet)] focus:outline-none"
          >
            <option value="">All Ranks</option>
            {ALL_RANKS.map(r => <option key={r} value={r}>{r.charAt(0) + r.slice(1).toLowerCase()}</option>)}
          </select>
          {(region || language || rank) && (
            <button
              onClick={() => { setRegion(''); setLanguage(''); setRank(''); setPage(1) }}
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
          {filteredData?.map((t, i) => (
            <TranslationCard
              key={t.id}
              t={t}
              index={(page - 1) * 25 + i}
              onVote={(id, dir) => voteMutation.mutate({ id, direction: dir })}
            />
          ))}

          {(!filteredData || filteredData.length === 0) && !isLoading && (
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
