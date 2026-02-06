import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'

interface FeedbackItem {
  id: number
  translation_id: number
  username: string
  translation: string
  feedback_text: string
  created_at: string
}

interface FeedbackResponse {
  data: FeedbackItem[]
  pagination: { page: number; limit: number; total: number }
}

async function fetchFeedback(page: number, credentials: string): Promise<FeedbackResponse> {
  const res = await fetch(`/admin/feedback?page=${page}&limit=25`, {
    headers: { Authorization: `Basic ${credentials}` },
  })
  if (res.status === 401) throw new Error('Unauthorized')
  if (!res.ok) throw new Error('Failed to fetch feedback')
  return res.json()
}

export function Admin() {
  const [password, setPassword] = useState('')
  const [credentials, setCredentials] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading, error } = useQuery({
    queryKey: ['admin-feedback', page, credentials],
    queryFn: () => fetchFeedback(page, credentials),
    enabled: credentials !== '',
    retry: false,
  })

  if (!credentials) {
    return (
      <div className="max-w-md mx-auto py-16">
        <div className="pixel-border bg-[var(--card)] pixel-shadow-lg p-8">
          <h1 className="pixel-font text-sm tracking-wide mb-6">Admin Login</h1>
          <form onSubmit={e => { e.preventDefault(); setCredentials(btoa('admin:' + password)) }}>
            <input
              type="password"
              value={password}
              onChange={e => setPassword(e.target.value)}
              placeholder="Admin password"
              className="w-full bg-[var(--muted)] border-4 border-[var(--border)] rounded-[8px] px-4 py-3 text-sm mb-4 focus:border-[var(--violet)] focus:outline-none"
            />
            <button
              type="submit"
              className="pixel-font text-xs px-6 py-2 bg-[var(--violet)] text-white border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm tracking-wide uppercase hover:bg-[var(--violet-deep)] transition-all duration-150 btn-press w-full"
            >
              Login
            </button>
          </form>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="max-w-md mx-auto py-16">
        <div className="pixel-border bg-[var(--card)] pixel-shadow-lg p-8 text-center">
          <p className="pixel-font text-sm text-[var(--destructive)] mb-4">Invalid password</p>
          <button
            onClick={() => { setCredentials(''); setPassword('') }}
            className="pixel-font text-xs px-4 py-2 bg-[var(--muted)] border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm tracking-wide uppercase btn-press"
          >
            Try Again
          </button>
        </div>
      </div>
    )
  }

  const totalPages = data ? Math.ceil(data.pagination.total / 25) : 0

  return (
    <div className="max-w-4xl mx-auto py-8 px-6">
      <div className="flex items-center justify-between mb-8">
        <h1 className="pixel-font text-lg tracking-wide">Feedback ({data?.pagination.total ?? 0})</h1>
        <button
          onClick={() => { setCredentials(''); setPassword('') }}
          className="pixel-font text-[10px] px-3 py-1.5 bg-[var(--muted)] border-4 border-[var(--border)] rounded-[8px] pixel-shadow-sm tracking-wide uppercase btn-press"
        >
          Logout
        </button>
      </div>

      {isLoading ? (
        <div className="pixel-border bg-[var(--card)] pixel-shadow-md p-12 text-center">
          <p className="pixel-font text-sm text-[var(--foreground-muted)]">Loading...</p>
        </div>
      ) : (
        <div className="space-y-3">
          {data?.data.map(fb => (
            <div key={fb.id} className="pixel-border bg-[var(--card)] pixel-shadow-md p-4">
              <div className="flex items-baseline gap-3 mb-2">
                <span className="pixel-font text-sm text-[var(--primary)]">{fb.username}</span>
                <span className="text-[var(--border)]">&rarr;</span>
                <span className="font-bold text-sm">{fb.translation}</span>
              </div>
              <div className="bg-[var(--muted)] border-2 border-[var(--border-light)] rounded-[4px] p-3 mb-2">
                <p className="text-sm">{fb.feedback_text}</p>
              </div>
              <p className="mono-font text-[10px] text-[var(--foreground-muted)] tracking-wide" style={{ opacity: 0.6 }}>
                {new Date(fb.created_at).toLocaleString()}
              </p>
            </div>
          ))}

          {data?.data.length === 0 && (
            <div className="pixel-border bg-[var(--card)] pixel-shadow-md p-8 text-center">
              <p className="pixel-font text-sm text-[var(--foreground-muted)]">No feedback yet</p>
            </div>
          )}
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-4 mt-6">
          <button
            disabled={page <= 1}
            onClick={() => setPage(p => p - 1)}
            className="pixel-font text-xs px-4 py-2 pixel-border bg-[var(--card)] pixel-shadow-sm disabled:opacity-40 btn-press tracking-wide"
          >
            &larr; Prev
          </button>
          <span className="mono-font text-xs tracking-widest text-[var(--foreground-muted)]">
            {page} / {totalPages}
          </span>
          <button
            disabled={page >= totalPages}
            onClick={() => setPage(p => p + 1)}
            className="pixel-font text-xs px-4 py-2 pixel-border bg-[var(--card)] pixel-shadow-sm disabled:opacity-40 btn-press tracking-wide"
          >
            Next &rarr;
          </button>
        </div>
      )}
    </div>
  )
}
