import { z } from 'zod'

export const translationSchema = z.object({
  id: z.number(),
  username: z.string(),
  translation: z.string(),
  explanation: z.string().nullable(),
  language: z.string(),
  region: z.string(),
  riot_verified: z.boolean(),
  upvotes: z.number(),
  downvotes: z.number(),
  created_at: z.string(),
})

export type Translation = z.infer<typeof translationSchema>

export const paginationSchema = z.object({
  page: z.number(),
  limit: z.number(),
  total: z.number(),
})

export const translationListResponseSchema = z.object({
  data: z.array(translationSchema),
  pagination: paginationSchema,
})

export type TranslationListResponse = z.infer<typeof translationListResponseSchema>

export const voteRequestSchema = z.object({
  vote: z.union([z.literal(1), z.literal(-1)]),
})

export const feedbackRequestSchema = z.object({
  text: z.string().min(1).max(500),
})

export const feedbackSchema = z.object({
  id: z.number(),
  translation_id: z.number(),
  username: z.string(),
  translation: z.string(),
  feedback_text: z.string(),
  created_at: z.string(),
})

export type Feedback = z.infer<typeof feedbackSchema>

export type SortOption = 'hot' | 'new' | 'top'
export type PeriodOption = 'hour' | 'day' | 'week' | 'month' | 'year' | 'all'
