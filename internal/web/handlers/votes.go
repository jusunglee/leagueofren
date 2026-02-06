package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jusunglee/leagueofren/internal/db"
	"github.com/jusunglee/leagueofren/internal/web/middleware"
)

const (
	visitorCookieName = "lor_visitor"
	// Vote uniqueness is per browser (cookie), not per IP. This is because multiple
	// users can share an IP due to NAT (e.g. a LAN party, office, university dorm).
	// With IP-only uniqueness, an entire household would be limited to one vote per
	// translation. Instead, we use a random UUID cookie as the visitor fingerprint,
	// with a per-IP cap of 10 votes as an abuse backstop.
	maxVotesPerIP = 10
)

type VoteHandler struct {
	repo db.Repository
	log  *slog.Logger
}

func NewVoteHandler(repo db.Repository, log *slog.Logger) *VoteHandler {
	return &VoteHandler{repo: repo, log: log}
}

type voteRequest struct {
	Vote int16 `json:"vote"`
}

type voteResponse struct {
	Upvotes   int32 `json:"upvotes"`
	Downvotes int32 `json:"downvotes"`
}

func (h *VoteHandler) Vote(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	translationID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req voteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Vote != 1 && req.Vote != -1 {
		writeError(w, http.StatusBadRequest, "vote must be 1 or -1")
		return
	}

	ipHash := hashIP(middleware.ClientIP(r))
	visitorID := getOrSetVisitorCookie(w, r)

	// Check IP-level rate limit before processing the vote
	ipCount, err := h.repo.CountVotesByIP(r.Context(), ipHash)
	if err != nil {
		h.log.ErrorContext(r.Context(), "counting votes by IP", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if ipCount >= maxVotesPerIP {
		writeError(w, http.StatusTooManyRequests, "too many votes from this network")
		return
	}

	err = h.repo.WithTx(r.Context(), func(txRepo db.Repository) error {
		existing, err := txRepo.GetVote(r.Context(), db.GetVoteParams{
			TranslationID: translationID,
			VisitorID:     visitorID,
		})

		if db.IsNoRows(err) {
			if _, err := txRepo.UpsertVote(r.Context(), db.UpsertVoteParams{
				TranslationID: translationID,
				IpHash:        ipHash,
				VisitorID:     visitorID,
				Vote:          req.Vote,
			}); err != nil {
				return fmt.Errorf("inserting vote: %w", err)
			}
			if req.Vote == 1 {
				return txRepo.IncrementUpvotes(r.Context(), translationID)
			}
			return txRepo.IncrementDownvotes(r.Context(), translationID)
		}
		if err != nil {
			return fmt.Errorf("getting existing vote: %w", err)
		}

		if existing.Vote == req.Vote {
			if _, err := txRepo.DeleteVote(r.Context(), db.DeleteVoteParams{
				TranslationID: translationID,
				VisitorID:     visitorID,
			}); err != nil {
				return fmt.Errorf("deleting vote: %w", err)
			}
			if req.Vote == 1 {
				return txRepo.DecrementUpvotes(r.Context(), translationID)
			}
			return txRepo.DecrementDownvotes(r.Context(), translationID)
		}

		if _, err := txRepo.UpsertVote(r.Context(), db.UpsertVoteParams{
			TranslationID: translationID,
			IpHash:        ipHash,
			VisitorID:     visitorID,
			Vote:          req.Vote,
		}); err != nil {
			return fmt.Errorf("updating vote: %w", err)
		}
		if req.Vote == 1 {
			if err := txRepo.IncrementUpvotes(r.Context(), translationID); err != nil {
				return err
			}
			return txRepo.DecrementDownvotes(r.Context(), translationID)
		}
		if err := txRepo.IncrementDownvotes(r.Context(), translationID); err != nil {
			return err
		}
		return txRepo.DecrementUpvotes(r.Context(), translationID)
	})

	if err != nil {
		h.log.ErrorContext(r.Context(), "processing vote", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	t, err := h.repo.GetPublicTranslation(r.Context(), translationID)
	if err != nil {
		h.log.ErrorContext(r.Context(), "fetching translation after vote", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, voteResponse{
		Upvotes:   t.Upvotes,
		Downvotes: t.Downvotes,
	})
}

func hashIP(ip string) string {
	dailySalt := time.Now().Format("2006-01-02")
	h := sha256.Sum256([]byte(ip + dailySalt))
	return fmt.Sprintf("%x", h)
}

func getOrSetVisitorCookie(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(visitorCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	id := generateUUID()
	http.SetCookie(w, &http.Cookie{
		Name:     visitorCookieName,
		Value:    id,
		Path:     "/",
		MaxAge:   365 * 24 * 60 * 60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return id
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
