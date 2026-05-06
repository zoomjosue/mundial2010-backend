package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"mundial2010/internal/models"
	"net/http"
	"strings"
)

type RatingsHandler struct {
	DB *sql.DB
}

func NewRatingsHandler(db *sql.DB) *RatingsHandler {
	return &RatingsHandler{DB: db}
}

func (h *RatingsHandler) Create(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid serie ID", nil)
		return
	}

	var exists bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM series WHERE id=$1)", id).Scan(&exists)
	if !exists {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Serie with id %d not found", id), nil)
		return
	}

	var input models.RatingInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body", nil)
		return
	}

	details := map[string]string{}
	if input.Score < 1 || input.Score > 10 {
		details["score"] = "score must be between 1 and 10"
	}
	if len(details) > 0 {
		writeError(w, http.StatusBadRequest, "Validation failed", details)
		return
	}

	var rating models.Rating
	err = h.DB.QueryRow(
		`INSERT INTO ratings (serie_id, score, comment) VALUES ($1,$2,$3) RETURNING id, serie_id, score, comment, created_at`,
		id, input.Score, input.Comment,
	).Scan(&rating.ID, &rating.SerieID, &rating.Score, &rating.Comment, &rating.CreatedAt)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error saving rating", nil)
		return
	}

	writeJSON(w, http.StatusCreated, rating)
}

func (h *RatingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid serie ID", nil)
		return
	}

	var exists bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM series WHERE id=$1)", id).Scan(&exists)
	if !exists {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Serie with id %d not found", id), nil)
		return
	}

	rows, err := h.DB.Query(
		"SELECT id, serie_id, score, comment, created_at FROM ratings WHERE serie_id=$1 ORDER BY created_at DESC", id,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error fetching ratings", nil)
		return
	}
	defer rows.Close()

	ratings := []models.Rating{}
	var totalScore int
	for rows.Next() {
		var rt models.Rating
		if err := rows.Scan(&rt.ID, &rt.SerieID, &rt.Score, &rt.Comment, &rt.CreatedAt); err != nil {
			continue
		}
		ratings = append(ratings, rt)
		totalScore += rt.Score
	}

	avg := 0.0
	if len(ratings) > 0 {
		avg = float64(totalScore) / float64(len(ratings))
	}

	writeJSON(w, http.StatusOK, models.RatingSummary{
		SerieID: id,
		Average: math.Round(avg*10) / 10,
		Count:   len(ratings),
		Ratings: ratings,
	})
}

func (h *RatingsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		writeError(w, http.StatusBadRequest, "Invalid path", nil)
		return
	}

	var ratingID int
	fmt.Sscanf(parts[len(parts)-1], "%d", &ratingID)

	res, err := h.DB.Exec("DELETE FROM ratings WHERE id=$1", ratingID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error deleting rating", nil)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		writeError(w, http.StatusNotFound, "Rating not found", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
