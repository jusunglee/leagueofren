package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"
	"unicode"

	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/jusunglee/leagueofren/internal/translation"
	"github.com/jusunglee/leagueofren/internal/transliteration"
)

type TranslationHandler struct {
	repo       db.Repository
	log        *slog.Logger
	riot       *riot.DirectClient
	translator *translation.Translator
}

func NewTranslationHandler(repo db.Repository, log *slog.Logger, riotClient *riot.DirectClient, translator *translation.Translator) *TranslationHandler {
	return &TranslationHandler{repo: repo, log: log, riot: riotClient, translator: translator}
}

type translationResponse struct {
	ID              int64    `json:"id"`
	Username        string   `json:"username"`
	Transliteration string   `json:"transliteration"`
	Translation     string   `json:"translation"`
	Explanation     *string  `json:"explanation,omitempty"`
	Language        string   `json:"language"`
	Region          string   `json:"region"`
	RiotVerified    bool     `json:"riot_verified"`
	Rank            *string  `json:"rank,omitempty"`
	TopChampions    []string `json:"top_champions,omitempty"`
	Upvotes         int32    `json:"upvotes"`
	Downvotes       int32    `json:"downvotes"`
	Score           float64  `json:"score,omitempty"`
	CreatedAt       string   `json:"created_at"`
	FirstSeen       string   `json:"first_seen,omitempty"`
}

type paginationMeta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type listResponse struct {
	Data       []translationResponse `json:"data"`
	Pagination paginationMeta        `json:"pagination"`
}

func toTranslationResponse(t db.PublicTranslation) translationResponse {
	resp := translationResponse{
		ID:              t.ID,
		Username:        t.Username,
		Transliteration: transliteration.Transliterate(t.Username),
		Translation:     t.Translation,
		Language:        t.Language,
		Region:          t.Region,
		RiotVerified:    t.RiotVerified,
		Upvotes:         t.Upvotes,
		Downvotes:       t.Downvotes,
		CreatedAt:       t.CreatedAt.Format(time.RFC3339),
		FirstSeen:       t.FirstSeen.Format(time.RFC3339),
	}
	if t.Explanation.Valid {
		resp.Explanation = &t.Explanation.String
	}
	if t.Rank.Valid {
		resp.Rank = &t.Rank.String
	}
	if t.TopChampions.Valid && t.TopChampions.String != "" {
		var champs []string
		if err := json.Unmarshal([]byte(t.TopChampions.String), &champs); err == nil {
			resp.TopChampions = champs
		}
	}
	return resp
}

func hotScore(upvotes, downvotes int32, createdAt time.Time) float64 {
	diff := int64(upvotes) - int64(downvotes)
	absDiff := diff
	if absDiff < 0 {
		absDiff = -absDiff
	}
	magnitude := math.Log10(math.Max(float64(absDiff), 1))
	epochSeconds := float64(createdAt.Unix())
	return magnitude + epochSeconds/45000
}

func (h *TranslationHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	sort := q.Get("sort")
	if sort == "" {
		sort = "hot"
	}
	period := q.Get("period")
	if period == "" {
		period = "day"
	}
	region := q.Get("region")
	language := q.Get("language")

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 25
	}
	offset := (page - 1) * limit

	total, err := h.repo.CountPublicTranslations(r.Context(), db.CountPublicTranslationsParams{
		Region:   region,
		Language: language,
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "counting translations", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var translations []db.PublicTranslation

	switch sort {
	case "top":
		cutoff := periodCutoff(period)
		translations, err = h.repo.ListPublicTranslationsTop(r.Context(), db.ListPublicTranslationsTopParams{
			Region:    region,
			Language:  language,
			Limit:     int32(limit),
			Offset:    int32(offset),
			CreatedAt: cutoff,
		})
	case "hot":
		// Fetch from "new" and sort by hot score in Go
		translations, err = h.repo.ListPublicTranslationsNew(r.Context(), db.ListPublicTranslationsNewParams{
			Region:   region,
			Language: language,
			Limit:    500,
			Offset:   0,
		})
	default:
		translations, err = h.repo.ListPublicTranslationsNew(r.Context(), db.ListPublicTranslationsNewParams{
			Region:   region,
			Language: language,
			Limit:    int32(limit),
			Offset:   int32(offset),
		})
	}

	if err != nil {
		h.log.ErrorContext(r.Context(), "listing translations", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var data []translationResponse

	if sort == "hot" {
		type scored struct {
			resp  translationResponse
			score float64
		}
		items := make([]scored, len(translations))
		for i, t := range translations {
			s := hotScore(t.Upvotes, t.Downvotes, t.CreatedAt)
			resp := toTranslationResponse(t)
			resp.Score = s
			items[i] = scored{resp: resp, score: s}
		}
		// Sort descending by score (insertion sort is fine for <= 500 items)
		for i := 1; i < len(items); i++ {
			key := items[i]
			j := i - 1
			for j >= 0 && items[j].score < key.score {
				items[j+1] = items[j]
				j--
			}
			items[j+1] = key
		}

		start := offset
		end := offset + limit
		if start > len(items) {
			start = len(items)
		}
		if end > len(items) {
			end = len(items)
		}
		for _, item := range items[start:end] {
			data = append(data, item.resp)
		}
	} else {
		for _, t := range translations {
			data = append(data, toTranslationResponse(t))
		}
	}

	if data == nil {
		data = []translationResponse{}
	}

	writeJSON(w, http.StatusOK, listResponse{
		Data: data,
		Pagination: paginationMeta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}

func (h *TranslationHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	t, err := h.repo.GetPublicTranslation(r.Context(), id)
	if err != nil {
		if db.IsNoRows(err) {
			writeError(w, http.StatusNotFound, "translation not found")
			return
		}
		h.log.ErrorContext(r.Context(), "getting translation", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, toTranslationResponse(t))
}

type createTranslationRequest struct {
	Username string `json:"username"`
	Region   string `json:"region"`
}

func (h *TranslationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Username == "" || req.Region == "" {
		writeError(w, http.StatusBadRequest, "username and region are required")
		return
	}

	// Validate that this is a real Riot username (must include #tag)
	gameName, tagLine, err := riot.ParseRiotID(req.Username)
	if err != nil {
		writeError(w, http.StatusBadRequest, "username must include a tag (e.g. Player#NA1)")
		return
	}

	account, err := h.riot.GetAccountByRiotID(gameName, tagLine, req.Region)
	if err != nil {
		h.log.WarnContext(r.Context(), "riot lookup failed", "gameName", gameName, "tagLine", tagLine, "region", req.Region, "error", err)
		writeError(w, http.StatusBadRequest, "username not found on Riot servers")
		return
	}
	puuid := account.PUUID

	// Server-side translation via LLM (ignores any client-provided translation)
	translations, err := h.translator.TranslateUsernames(r.Context(), []string{gameName})
	if err != nil || len(translations) == 0 {
		h.log.WarnContext(r.Context(), "translation failed", "username", req.Username, "error", err)
		writeError(w, http.StatusInternalServerError, "translation failed")
		return
	}

	t := translations[0]
	language := detectLanguageFromName(gameName)

	var result db.PublicTranslation
	err = h.repo.WithTx(r.Context(), func(txRepo db.Repository) error {
		_, err := txRepo.UpsertPlayer(r.Context(), db.UpsertPlayerParams{
			Username: req.Username,
			Region:   req.Region,
			Puuid:    sql.NullString{String: puuid, Valid: puuid != ""},
		})
		if err != nil {
			return err
		}

		params := db.UpsertPublicTranslationParams{
			Username:       req.Username,
			Translation:    t.Translated,
			Language:       language,
			PlayerUsername: req.Username,
			RiotVerified:   tagLine != "",
		}
		if t.Explanation != "" {
			params.Explanation = sql.NullString{String: t.Explanation, Valid: true}
		}

		result, err = txRepo.UpsertPublicTranslation(r.Context(), params)
		return err
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "upserting translation", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	result.Region = req.Region

	writeJSON(w, http.StatusCreated, toTranslationResponse(result))
}

func detectLanguageFromName(name string) string {
	for _, r := range name {
		if unicode.Is(unicode.Hangul, r) {
			return "korean"
		}
	}
	return "chinese"
}

func periodCutoff(period string) time.Time {
	now := time.Now()
	switch period {
	case "hour":
		return now.Add(-1 * time.Hour)
	case "day":
		return now.Add(-24 * time.Hour)
	case "week":
		return now.Add(-7 * 24 * time.Hour)
	case "month":
		return now.Add(-30 * 24 * time.Hour)
	case "year":
		return now.Add(-365 * 24 * time.Hour)
	case "all":
		return time.Time{}
	default:
		return now.Add(-24 * time.Hour)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
