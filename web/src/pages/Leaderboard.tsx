import { useState, useRef, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ChevronUp, ChevronDown, MessageCircleQuestion, ChevronDown as ChevronDownIcon, X, MessageSquarePlus, Send, SlidersHorizontal } from 'lucide-react'
import { listTranslations, vote, submitFeedback, RateLimitError } from '../lib/api'
import type { SortOption, PeriodOption, Translation } from '../lib/schemas'
import type { TranslationListResponse } from '../lib/schemas'

// ── Filled SVG icons for sort tabs ──

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

function SortIcon({ type, filled, color }: { type: 'flame' | 'sparkles' | 'trophy'; filled: boolean; color: string }) {
  switch (type) {
    case 'flame': return <FlameIcon filled={filled} color={color} />
    case 'sparkles': return <SparklesIcon filled={filled} color={color} />
    case 'trophy': return <TrophyIcon filled={filled} color={color} />
  }
}

// ── Custom Pixel Dropdown with fuzzy search ──

interface DropdownOption {
  value: string
  label: string
  icon?: string
  image?: string
}

function PixelDropdown({ options, value, onChange, placeholder }: {
  options: DropdownOption[]
  value: string
  onChange: (v: string) => void
  placeholder: string
}) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const ref = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  useEffect(() => {
    if (open && inputRef.current) inputRef.current.focus()
  }, [open])

  const selected = options.find(o => o.value === value)
  const filtered = search
    ? options.filter(o => o.label.toLowerCase().includes(search.toLowerCase()))
    : options

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => { setOpen(!open); setSearch('') }}
        className="pixel-font text-[10px] px-3 py-2 pixel-border bg-[var(--card)] pixel-shadow-sm tracking-wide inline-flex items-center gap-2 hover:bg-[var(--muted)] transition-all duration-150 btn-press min-w-[140px] justify-between"
      >
        <span className="inline-flex items-center gap-1.5 truncate">
          {selected?.icon && <span className="text-sm">{selected.icon}</span>}
          {selected?.image && <img src={selected.image} alt="" className="w-5 h-5 object-contain" />}
          {selected ? selected.label : placeholder}
        </span>
        <ChevronDownIcon size={12} strokeWidth={3} className={`transition-transform duration-150 ${open ? 'rotate-180' : ''}`} />
      </button>

      {open && (
        <div className="absolute top-full mt-1 left-0 z-50 min-w-[180px] pixel-border bg-[var(--card)] pixel-shadow-lg overflow-hidden">
          <div className="p-2 border-b-2 border-[var(--border-light)]">
            <input
              ref={inputRef}
              type="text"
              value={search}
              onChange={e => setSearch(e.target.value)}
              placeholder="Search..."
              className="w-full bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] px-2 py-1 text-base focus:border-[var(--violet)] focus:outline-none"
            />
          </div>
          <div className="max-h-[240px] overflow-y-auto">
            <button
              onClick={() => { onChange(''); setOpen(false) }}
              className={`w-full text-left px-3 py-1.5 text-xs flex items-center gap-2 hover:bg-[var(--muted)] transition-colors ${!value ? 'bg-[var(--muted)] font-bold' : ''}`}
            >
              {placeholder}
            </button>
            {filtered.map(opt => (
              <button
                key={opt.value}
                onClick={() => { onChange(opt.value); setOpen(false) }}
                className={`w-full text-left px-3 py-1.5 text-xs flex items-center gap-2 hover:bg-[var(--muted)] transition-colors ${value === opt.value ? 'bg-[var(--muted)] font-bold' : ''}`}
              >
                {opt.icon && <span className="text-sm">{opt.icon}</span>}
                {opt.image && <img src={opt.image} alt="" className="w-5 h-5 object-contain" />}
                {opt.label}
              </button>
            ))}
            {filtered.length === 0 && (
              <div className="px-3 py-2 text-xs text-[var(--foreground-muted)]">No matches</div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

// ── Constants ──

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

const REGION_EMOJI: Record<string, string> = {
  NA: '🇺🇸', EUW: '🇪🇺', EUNE: '🇪🇺', KR: '🇰🇷', JP: '🇯🇵',
  BR: '🇧🇷', LAN: '🌎', LAS: '🌎', OCE: '🇦🇺', TR: '🇹🇷', RU: '🇷🇺', TW: '🇹🇼',
}

const LANGUAGE_EMOJI: Record<string, string> = { korean: '🇰🇷', chinese: '🇨🇳' }

const RANK_ICON: Record<string, string> = {
  IRON: '/iron.png',
  BRONZE: '/bronze.png',
  SILVER: '/silver.png',
  GOLD: '/gold.png',
  PLATINUM: '/platinum.png',
  EMERALD: '/emerald.png',
  DIAMOND: '/diamond.png',
  MASTER: '/master.png',
  GRANDMASTER: '/grandmaster.png',
  CHALLENGER: '/challenger.png',
}

const REGION_OPTIONS: DropdownOption[] = ['NA', 'EUW', 'EUNE', 'KR', 'JP', 'BR', 'LAN', 'LAS', 'OCE', 'TR', 'RU', 'TW']
  .map(r => ({ value: r, label: r, icon: REGION_EMOJI[r] }))

const LANGUAGE_OPTIONS: DropdownOption[] = [
  { value: 'korean', label: 'Korean', icon: '🇰🇷' },
  { value: 'chinese', label: 'Chinese', icon: '🇨🇳' },
]

const RANK_OPTIONS: DropdownOption[] = Object.entries(RANK_ICON)
  .map(([rank, url]) => ({ value: rank, label: rank.charAt(0) + rank.slice(1).toLowerCase(), image: url }))

const BADGE_COLORS = [
  'var(--primary)', 'var(--violet)', 'var(--accent)',
  'var(--sky)', 'var(--lavender)', 'var(--secondary)',
]

// ── Sub-components ──

function ListRankBadge({ index }: { index: number }) {
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

function timeAgo(dateStr: string): string {
  const seconds = Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000)
  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  const months = Math.floor(days / 30)
  if (months < 12) return `${months}mo ago`
  const years = Math.floor(days / 365)
  return `${years}y ago`
}

function buildLearnMoreUrl(username: string, translation: string, explanation: string | null) {
  const prompt = `Tell me more about the League of Legends summoner name "${username}". It translates to "${translation}"${explanation ? ` (${explanation})` : ''}. What's the cultural context, any references to games, anime, literature, or memes? Is this a common naming pattern?`
  return `https://chatgpt.com/?q=${encodeURIComponent(prompt)}`
}

function buildOpggUrl(username: string, region: string) {
  return `https://www.op.gg/summoners/${region.toLowerCase()}/${encodeURIComponent(username)}`
}

function buildPorofessorUrl(username: string, region: string) {
  return `https://www.porofessor.gg/live/${region.toLowerCase()}/${encodeURIComponent(username)}`
}

function TranslationCard({ t, index, onVote, onFeedback, voteAnimation }: {
  t: Translation
  index: number
  onVote: (id: number, dir: 1 | -1) => void
  onFeedback: (id: number, text: string) => void
  voteAnimation?: 'up' | 'down' | 'shake' | null
}) {
  const [showFeedback, setShowFeedback] = useState(false)
  const [feedbackText, setFeedbackText] = useState('')
  const [feedbackSent, setFeedbackSent] = useState(false)
  const score = t.upvotes - t.downvotes
  const isTop3 = index < 3
  const rank = t.rank?.toUpperCase()

  const voteColClass = voteAnimation === 'shake' ? 'animate-vote-shake' : ''
  const scoreAnimClass = voteAnimation === 'up' ? 'animate-vote-up' : voteAnimation === 'down' ? 'animate-vote-down' : ''
  const upChevronFlash = voteAnimation === 'up' ? 'text-[var(--secondary)]' : ''
  const downChevronFlash = voteAnimation === 'down' ? 'text-[var(--destructive)]' : ''

  return (
    <div
      className={`relative pixel-border bg-[var(--card)] p-4 lg:p-5 flex items-center gap-4 transition-all duration-200 animate-fade-in ${
        isTop3
          ? 'pixel-border-double bg-[var(--background-alt)] shadow-[6px_6px_0px_0px_var(--violet)] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[8px_8px_0px_0px_var(--violet)]'
          : 'pixel-shadow-md hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[6px_6px_0px_0px_var(--border)]'
      }`}
      style={{ animationDelay: `${index * 40}ms` }}
    >
      <ListRankBadge index={index} />

      <div className={`flex flex-col items-center gap-0 min-w-[48px] ${voteColClass}`}>
        <button onClick={() => onVote(t.id, 1)} className={`w-10 h-10 flex items-center justify-center border-2 border-transparent rounded-[4px] hover:border-[var(--secondary)] hover:bg-[var(--muted)] transition-all duration-150 focus-visible:ring-2 focus-visible:ring-[var(--ring)] ${upChevronFlash}`} aria-label="Upvote">
          <ChevronUp size={24} strokeWidth={3} />
        </button>
        <span className={`pixel-font text-sm leading-none py-1 ${scoreAnimClass} ${score > 0 ? 'text-[var(--secondary)]' : score < 0 ? 'text-[var(--destructive)]' : 'text-[var(--foreground-muted)]'}`}>
          {score}
        </span>
        <button onClick={() => onVote(t.id, -1)} className={`w-10 h-10 flex items-center justify-center border-2 border-transparent rounded-[4px] hover:border-[var(--destructive)] hover:bg-[var(--muted)] transition-all duration-150 focus-visible:ring-2 focus-visible:ring-[var(--ring)] ${downChevronFlash}`} aria-label="Downvote">
          <ChevronDown size={24} strokeWidth={3} />
        </button>
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-baseline gap-3 flex-wrap">
          <span className="pixel-font text-base lg:text-lg text-[var(--primary)] tracking-wide">{t.username}</span>
          <span className="text-[var(--border)] font-bold">&rarr;</span>
          <span className="font-bold text-base lg:text-lg">{t.translation}</span>
        </div>
        {t.explanation && (
          <p className="text-sm text-[var(--foreground-muted)] mt-1 leading-relaxed">{t.explanation}</p>
        )}
        {t.first_seen && (
          <p className="mono-font text-[10px] text-[var(--foreground-muted)] mt-1 tracking-wide" style={{ opacity: 0.6 }}>
            first seen {timeAgo(t.first_seen)}
          </p>
        )}
        <div className="flex items-center gap-2 mt-3 flex-wrap">
          <span className="mono-font text-xs px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1">
            <span className="text-sm leading-none" aria-hidden="true">{REGION_EMOJI[t.region] || '🌍'}</span>
            {t.region}
          </span>
          <span className="mono-font text-xs px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1">
            <span className="text-sm leading-none" aria-hidden="true">{LANGUAGE_EMOJI[t.language] || '🌐'}</span>
            {t.language}
          </span>
          {rank && RANK_ICON[rank] && (
            <span className="mono-font text-xs px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1">
              <img src={RANK_ICON[rank]} alt="" className="w-4 h-4 object-contain" aria-hidden="true" />
              {rank}
            </span>
          )}
          {t.top_champions?.map(champ => (
            <span key={champ} className="mono-font text-xs px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1">
              <span className="text-sm leading-none" aria-hidden="true">⚔️</span>
              {champ}
            </span>
          ))}

          <span className="hidden lg:inline w-px h-4 bg-[var(--border-light)]" />

          <a href={buildOpggUrl(t.username, t.region)} target="_blank" rel="noopener noreferrer"
            className="mono-font text-[10px] px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1 hover:bg-[#5383E8] hover:text-white hover:border-[var(--border)] transition-all duration-150">
            <OpggIcon /> OP.GG
          </a>
          <a href={buildPorofessorUrl(t.username, t.region)} target="_blank" rel="noopener noreferrer"
            className="mono-font text-[10px] px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1 hover:bg-[#785A28] hover:text-white hover:border-[var(--border)] transition-all duration-150">
            <PorofessorIcon /> Porofessor
          </a>
          <a href={buildLearnMoreUrl(t.username, t.translation, t.explanation)} target="_blank" rel="noopener noreferrer"
            className="mono-font text-[10px] px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1 hover:bg-[var(--accent)] hover:text-[var(--background)] hover:border-[var(--border)] transition-all duration-150">
            <MessageCircleQuestion size={10} strokeWidth={2.5} /> Learn More
          </a>
          <button
            onClick={() => { setShowFeedback(!showFeedback); setFeedbackSent(false) }}
            className="mono-font text-[10px] px-2 py-0.5 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] tracking-widest uppercase inline-flex items-center gap-1 hover:bg-[var(--violet)] hover:text-white hover:border-[var(--border)] transition-all duration-150"
          >
            <MessageSquarePlus size={10} strokeWidth={2.5} /> Feedback
          </button>
        </div>
        {showFeedback && (
          <div className="mt-3 flex items-center gap-2">
            {feedbackSent ? (
              <span className="pixel-font text-[10px] text-[var(--secondary)] tracking-wide">Thanks for your feedback!</span>
            ) : (
              <>
                <input
                  type="text"
                  value={feedbackText}
                  onChange={e => setFeedbackText(e.target.value)}
                  onKeyDown={e => {
                    if (e.key === 'Enter' && feedbackText.trim()) {
                      onFeedback(t.id, feedbackText.trim())
                      setFeedbackText('')
                      setFeedbackSent(true)
                    }
                  }}
                  placeholder="Suggest a correction..."
                  maxLength={500}
                  className="flex-1 bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] px-3 py-1.5 text-xs focus:border-[var(--violet)] focus:outline-none"
                />
                <button
                  onClick={() => {
                    if (feedbackText.trim()) {
                      onFeedback(t.id, feedbackText.trim())
                      setFeedbackText('')
                      setFeedbackSent(true)
                    }
                  }}
                  disabled={!feedbackText.trim()}
                  className="pixel-font text-[10px] px-3 py-1.5 bg-[var(--violet)] text-white border-2 border-[var(--border)] rounded-[4px] pixel-shadow-sm tracking-wide uppercase hover:bg-[var(--violet-deep)] transition-all duration-150 btn-press disabled:opacity-40 inline-flex items-center gap-1"
                >
                  <Send size={10} strokeWidth={2.5} /> Send
                </button>
              </>
            )}
          </div>
        )}
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

// ── Pagination ──

function Pagination({ page, totalPages, setPage }: { page: number; totalPages: number; setPage: (fn: (p: number) => number) => void }) {
  if (totalPages <= 1) return null
  return (
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
  )
}

// ── Main Component ──

export function Leaderboard() {
  const queryClient = useQueryClient()
  const [sort, setSort] = useState<SortOption>('hot')
  const [period, setPeriod] = useState<PeriodOption>('week')
  const [region, setRegion] = useState('')
  const [language, setLanguage] = useState('')
  const [rank, setRank] = useState('')
  const [champion, setChampion] = useState('')
  const [page, setPage] = useState(1)
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [voteAnimations, setVoteAnimations] = useState<Record<number, 'up' | 'down' | 'shake'>>({})

  const queryKey = ['translations', sort, period, region, language, page]

  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: () => listTranslations({ sort, period, region, language, page, limit: 25 }),
    refetchInterval: 60_000,
  })

  const triggerVoteAnimation = (id: number, type: 'up' | 'down' | 'shake') => {
    setVoteAnimations(prev => ({ ...prev, [id]: type }))
    setTimeout(() => {
      setVoteAnimations(prev => {
        const next = { ...prev }
        delete next[id]
        return next
      })
    }, type === 'shake' ? 500 : 400)
  }

  const voteMutation = useMutation({
    mutationFn: ({ id, direction }: { id: number; direction: 1 | -1 }) => vote(id, direction),
    onMutate: async ({ id, direction }) => {
      await queryClient.cancelQueries({ queryKey })
      const previous = queryClient.getQueryData<TranslationListResponse>(queryKey)
      queryClient.setQueryData<TranslationListResponse>(queryKey, old => {
        if (!old) return old
        return {
          ...old,
          data: old.data.map(t =>
            t.id === id
              ? { ...t, upvotes: t.upvotes + (direction === 1 ? 1 : 0), downvotes: t.downvotes + (direction === -1 ? 1 : 0) }
              : t
          ),
        }
      })
      triggerVoteAnimation(id, direction === 1 ? 'up' : 'down')
      return { previous }
    },
    onError: (err, { id }, context) => {
      if (context?.previous) {
        queryClient.setQueryData(queryKey, context.previous)
      }
      if (err instanceof RateLimitError) {
        triggerVoteAnimation(id, 'shake')
      }
    },
    onSettled: () => {
      setTimeout(() => {
        queryClient.invalidateQueries({ queryKey: ['translations'] })
      }, 2000)
    },
  })

  const feedbackMutation = useMutation({
    mutationFn: ({ id, text }: { id: number; text: string }) => submitFeedback(id, text),
  })

  const filteredData = data?.data.filter(t => {
    if (rank && t.rank?.toUpperCase() !== rank) return false
    if (champion && !t.top_champions?.includes(champion)) return false
    return true
  }) ?? undefined

  // Build champion options from loaded data
  const championOptions: DropdownOption[] = (() => {
    if (!data?.data) return []
    const champs = new Set<string>()
    for (const t of data.data) {
      t.top_champions?.forEach(c => champs.add(c))
    }
    return Array.from(champs).sort().map(c => ({ value: c, label: c, icon: '⚔️' }))
  })()

  const totalPages = data ? Math.ceil(data.pagination.total / 25) : 0
  const hasFilters = region || language || rank || champion

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
              <SortIcon type={opt.icon} filled={isActive} color={opt.color} />
              {opt.label}
            </button>
          )
        })}

        <div className="flex-1" />

        {/* Desktop: inline filters */}
        <div className="hidden lg:flex flex-wrap items-center gap-3">
          {sort === 'top' && (
            <PixelDropdown
              options={PERIOD_OPTIONS.map(p => ({ value: p.value, label: p.label }))}
              value={period}
              onChange={v => { setPeriod(v as PeriodOption); setPage(1) }}
              placeholder="Period"
            />
          )}
          <PixelDropdown options={REGION_OPTIONS} value={region} onChange={v => { setRegion(v); setPage(1) }} placeholder="All Regions" />
          <PixelDropdown options={LANGUAGE_OPTIONS} value={language} onChange={v => { setLanguage(v); setPage(1) }} placeholder="All Languages" />
          <PixelDropdown options={RANK_OPTIONS} value={rank} onChange={v => { setRank(v); setPage(1) }} placeholder="All Ranks" />
          <PixelDropdown options={championOptions} value={champion} onChange={v => { setChampion(v); setPage(1) }} placeholder="All Champions" />
          {hasFilters && (
            <button
              onClick={() => { setRegion(''); setLanguage(''); setRank(''); setChampion(''); setPage(1) }}
              className="pixel-font text-[10px] px-3 py-2 bg-[var(--destructive)] text-white border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm tracking-wide uppercase hover:bg-[#b83a30] hover:-translate-x-0.5 hover:-translate-y-0.5 transition-all duration-150 btn-press inline-flex items-center gap-1"
            >
              <X size={12} strokeWidth={3} />
              Clear
            </button>
          )}
        </div>

        {/* Mobile: filter toggle icon */}
        <button
          onClick={() => setFiltersOpen(!filtersOpen)}
          className={`lg:hidden p-2 pixel-border transition-all duration-150 btn-press inline-flex items-center justify-center focus-visible:ring-2 focus-visible:ring-[var(--ring)] ${
            filtersOpen || hasFilters
              ? 'bg-[var(--violet)] text-white pixel-shadow-sm'
              : 'bg-[var(--card)] pixel-shadow-sm hover:bg-[var(--muted)] hover:-translate-x-0.5 hover:-translate-y-0.5 hover:shadow-[4px_4px_0px_0px_var(--border)]'
          }`}
          aria-label="Toggle filters"
        >
          <SlidersHorizontal size={16} strokeWidth={2.5} />
        </button>
      </div>

      {/* Mobile: collapsible filters */}
      {filtersOpen && (
        <div className="flex lg:hidden flex-wrap items-center gap-3 animate-fade-in">
          {sort === 'top' && (
            <PixelDropdown
              options={PERIOD_OPTIONS.map(p => ({ value: p.value, label: p.label }))}
              value={period}
              onChange={v => { setPeriod(v as PeriodOption); setPage(1) }}
              placeholder="Period"
            />
          )}
          <PixelDropdown options={REGION_OPTIONS} value={region} onChange={v => { setRegion(v); setPage(1) }} placeholder="All Regions" />
          <PixelDropdown options={LANGUAGE_OPTIONS} value={language} onChange={v => { setLanguage(v); setPage(1) }} placeholder="All Languages" />
          <PixelDropdown options={RANK_OPTIONS} value={rank} onChange={v => { setRank(v); setPage(1) }} placeholder="All Ranks" />
          <PixelDropdown options={championOptions} value={champion} onChange={v => { setChampion(v); setPage(1) }} placeholder="All Champions" />
          {hasFilters && (
            <button
              onClick={() => { setRegion(''); setLanguage(''); setRank(''); setChampion(''); setPage(1) }}
              className="pixel-font text-[10px] px-3 py-2 bg-[var(--destructive)] text-white border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm tracking-wide uppercase hover:bg-[#b83a30] hover:-translate-x-0.5 hover:-translate-y-0.5 transition-all duration-150 btn-press inline-flex items-center gap-1"
            >
              <X size={12} strokeWidth={3} />
              Clear
            </button>
          )}
        </div>
      )}

      {/* Pagination — top */}
      <Pagination page={page} totalPages={totalPages} setPage={setPage} />

      {/* Translation Cards — scrollable container */}
      <div className="overflow-y-auto max-h-[70vh] pixel-border bg-[var(--background-alt)] p-3 space-y-3">
        {isLoading ? (
          <div className="pixel-border bg-[var(--card)] pixel-shadow-md p-12 text-center">
            <p className="pixel-font text-sm text-[var(--foreground-muted)] tracking-wide">Loading translations...</p>
          </div>
        ) : (
          <>
            {filteredData?.map((t, i) => (
              <TranslationCard
                key={t.id}
                t={t}
                index={(page - 1) * 25 + i}
                onVote={(id, dir) => voteMutation.mutate({ id, direction: dir })}
                onFeedback={(id, text) => feedbackMutation.mutate({ id, text })}
                voteAnimation={voteAnimations[t.id]}
              />
            ))}

            {(!filteredData || filteredData.length === 0) && !isLoading && (
              <div className="pixel-border-double bg-[var(--background-alt)] pixel-shadow-violet p-8 lg:p-12 text-center">
                <p className="pixel-font text-sm text-[var(--foreground-muted)] mb-2">No translations found!</p>
                <p className="text-sm text-[var(--foreground-muted)]">Try adjusting your filters or check back later.</p>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
