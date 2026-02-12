package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/web/middleware"
)

type FeedbackHandler struct {
	repo db.Repository
	log  *slog.Logger
}

func NewFeedbackHandler(repo db.Repository, log *slog.Logger) *FeedbackHandler {
	return &FeedbackHandler{repo: repo, log: log}
}

type createFeedbackRequest struct {
	Text string `json:"text"`
}

type feedbackResponse struct {
	ID            int64  `json:"id"`
	TranslationID int64  `json:"translation_id"`
	FeedbackText  string `json:"feedback_text"`
	CreatedAt     string `json:"created_at"`
}

type adminFeedbackRow struct {
	ID            int64  `json:"id"`
	TranslationID int64  `json:"translation_id"`
	FeedbackText  string `json:"feedback_text"`
	Username      string `json:"username"`
	Translation   string `json:"translation"`
	CreatedAt     string `json:"created_at"`
}

func (h *FeedbackHandler) Create(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	translationID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req createFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Text == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}
	if len(req.Text) > 500 {
		writeError(w, http.StatusBadRequest, "feedback text must be 500 characters or fewer")
		return
	}

	ipHash := hashIP(middleware.ClientIP(r))

	fb, err := h.repo.CreatePublicFeedback(r.Context(), db.CreatePublicFeedbackParams{
		TranslationID: translationID,
		IpHash:        ipHash,
		FeedbackText:  req.Text,
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "creating feedback", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, feedbackResponse{
		ID:            fb.ID,
		TranslationID: fb.TranslationID,
		FeedbackText:  fb.FeedbackText,
		CreatedAt:     fb.CreatedAt.Format(time.RFC3339),
	})
}

func (h *FeedbackHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 25
	}
	offset := (page - 1) * limit

	total, err := h.repo.CountPublicFeedback(r.Context())
	if err != nil {
		h.log.ErrorContext(r.Context(), "counting feedback", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	rows, err := h.repo.ListPublicFeedback(r.Context(), db.ListPublicFeedbackParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		h.log.ErrorContext(r.Context(), "listing feedback", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	data := make([]adminFeedbackRow, len(rows))
	for i, row := range rows {
		data[i] = adminFeedbackRow{
			ID:            row.ID,
			TranslationID: row.TranslationID,
			FeedbackText:  row.FeedbackText,
			Username:      row.Username,
			Translation:   row.Translation,
			CreatedAt:     row.CreatedAt.Format(time.RFC3339),
		}
	}

	writeJSON(w, http.StatusOK, struct {
		Data       []adminFeedbackRow `json:"data"`
		Pagination paginationMeta     `json:"pagination"`
	}{
		Data: data,
		Pagination: paginationMeta{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}
