import type { TranslationListResponse, Translation, SortOption, PeriodOption } from './schemas'

export class RateLimitError extends Error {
  constructor() {
    super('Rate limited')
    this.name = 'RateLimitError'
  }
}

const API_BASE = '/api/v1'

interface ListParams {
  sort?: SortOption
  period?: PeriodOption
  region?: string
  language?: string
  page?: number
  limit?: number
}

export async function listTranslations(params: ListParams = {}): Promise<TranslationListResponse> {
  const searchParams = new URLSearchParams()
  if (params.sort) searchParams.set('sort', params.sort)
  if (params.period) searchParams.set('period', params.period)
  if (params.region) searchParams.set('region', params.region)
  if (params.language) searchParams.set('language', params.language)
  if (params.page) searchParams.set('page', String(params.page))
  if (params.limit) searchParams.set('limit', String(params.limit))

  const res = await fetch(`${API_BASE}/translations?${searchParams}`)
  if (!res.ok) throw new Error('Failed to fetch translations')
  return res.json()
}

export async function getTranslation(id: number): Promise<Translation> {
  const res = await fetch(`${API_BASE}/translations/${id}`)
  if (!res.ok) throw new Error('Failed to fetch translation')
  return res.json()
}

export async function vote(translationId: number, direction: 1 | -1): Promise<{ upvotes: number; downvotes: number }> {
  const res = await fetch(`${API_BASE}/translations/${translationId}/vote`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ vote: direction }),
  })
  if (res.status === 429) throw new RateLimitError()
  if (!res.ok) throw new Error('Failed to vote')
  return res.json()
}

export async function submitFeedback(translationId: number, text: string): Promise<void> {
  const res = await fetch(`${API_BASE}/translations/${translationId}/feedback`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text }),
  })
  if (!res.ok) throw new Error('Failed to submit feedback')
}
