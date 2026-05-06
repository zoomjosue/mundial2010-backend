package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"mundial2010/internal/models"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type SeriesHandler struct {
	DB *sql.DB
}

func NewSeriesHandler(db *sql.DB) *SeriesHandler {
	return &SeriesHandler{DB: db}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string, details map[string]string) {
	writeJSON(w, status, models.ErrorResponse{Error: msg, Details: details})
}

func (h *SeriesHandler) List(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	q := r.URL.Query().Get("q")
	sort := r.URL.Query().Get("sort")
	order := strings.ToUpper(r.URL.Query().Get("order"))

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	validSort := map[string]bool{"id": true, "name": true, "year": true, "genre": true, "created_at": true}
	if sort == "" || !validSort[sort] {
		sort = "id"
	}

	if order != "ASC" && order != "DESC" {
		order = "ASC"
	}

	whereClause := ""
	args := []interface{}{}
	argIdx := 1

	if q != "" {
		whereClause = fmt.Sprintf("WHERE name ILIKE $%d OR description ILIKE $%d OR genre ILIKE $%d", argIdx, argIdx, argIdx)
		args = append(args, "%"+q+"%")
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM series %s", whereClause)
	var total int
	if err := h.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		writeError(w, http.StatusInternalServerError, "Error counting series", nil)
		return
	}

	offset := (page - 1) * limit
	dataQuery := fmt.Sprintf(
		"SELECT id, name, description, genre, year, image_url, created_at, updated_at FROM series %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		whereClause, sort, order, argIdx, argIdx+1,
	)
	args = append(args, limit, offset)

	rows, err := h.DB.Query(dataQuery, args...)
	if err != nil {
		log.Println("Query error:", err)
		writeError(w, http.StatusInternalServerError, "Error fetching series", nil)
		return
	}
	defer rows.Close()

	series := []models.Serie{}
	for rows.Next() {
		var s models.Serie
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Genre, &s.Year, &s.ImageURL, &s.CreatedAt, &s.UpdatedAt); err != nil {
			continue
		}
		series = append(series, s)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	writeJSON(w, http.StatusOK, models.PaginatedSeries{
		Data:       series,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func (h *SeriesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID", nil)
		return
	}

	var s models.Serie
	err = h.DB.QueryRow(
		"SELECT id, name, description, genre, year, image_url, created_at, updated_at FROM series WHERE id = $1", id,
	).Scan(&s.ID, &s.Name, &s.Description, &s.Genre, &s.Year, &s.ImageURL, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Serie with id %d not found", id), nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error fetching serie", nil)
		return
	}

	writeJSON(w, http.StatusOK, s)
}

func (h *SeriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.SerieInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body", nil)
		return
	}

	details := validateSerie(input)
	if len(details) > 0 {
		writeError(w, http.StatusBadRequest, "Validation failed", details)
		return
	}

	var s models.Serie
	err := h.DB.QueryRow(
		`INSERT INTO series (name, description, genre, year, image_url) VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, name, description, genre, year, image_url, created_at, updated_at`,
		input.Name, input.Description, input.Genre, input.Year, input.ImageURL,
	).Scan(&s.ID, &s.Name, &s.Description, &s.Genre, &s.Year, &s.ImageURL, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		log.Println("Insert error:", err)
		writeError(w, http.StatusInternalServerError, "Error creating serie", nil)
		return
	}

	writeJSON(w, http.StatusCreated, s)
}

func (h *SeriesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID", nil)
		return
	}

	var input models.SerieInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body", nil)
		return
	}

	details := validateSerie(input)
	if len(details) > 0 {
		writeError(w, http.StatusBadRequest, "Validation failed", details)
		return
	}

	var s models.Serie
	err = h.DB.QueryRow(
		`UPDATE series SET name=$1, description=$2, genre=$3, year=$4, image_url=$5, updated_at=NOW()
		 WHERE id=$6 RETURNING id, name, description, genre, year, image_url, created_at, updated_at`,
		input.Name, input.Description, input.Genre, input.Year, input.ImageURL, id,
	).Scan(&s.ID, &s.Name, &s.Description, &s.Genre, &s.Year, &s.ImageURL, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Serie with id %d not found", id), nil)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error updating serie", nil)
		return
	}

	writeJSON(w, http.StatusOK, s)
}

func (h *SeriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID", nil)
		return
	}

	res, err := h.DB.Exec("DELETE FROM series WHERE id = $1", id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error deleting serie", nil)
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Serie with id %d not found", id), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SeriesHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid ID", nil)
		return
	}

	var exists bool
	h.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM series WHERE id=$1)", id).Scan(&exists)
	if !exists {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Serie with id %d not found", id), nil)
		return
	}

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "File too large. Maximum size is 1MB", nil)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Missing 'image' field in form", nil)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true}
	if !allowed[ext] {
		writeError(w, http.StatusBadRequest, "Invalid file type. Allowed: jpg, jpeg, png, webp, gif", nil)
		return
	}

	filename := fmt.Sprintf("%d_%d%s", id, time.Now().UnixMilli(), ext)
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, 0755)
	destPath := filepath.Join(uploadDir, filename)

	if err := saveFile(file, destPath); err != nil {
		writeError(w, http.StatusInternalServerError, "Error saving image", nil)
		return
	}

	imageURL := fmt.Sprintf("/uploads/%s", filename)
	h.DB.Exec("UPDATE series SET image_url=$1, updated_at=NOW() WHERE id=$2", imageURL, id)

	writeJSON(w, http.StatusOK, map[string]string{
		"image_url": imageURL,
		"message":   "Image uploaded successfully",
	})
}

func saveFile(src multipart.File, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

func validateSerie(input models.SerieInput) map[string]string {
	details := map[string]string{}
	if strings.TrimSpace(input.Name) == "" {
		details["name"] = "name is required"
	}
	if len(input.Name) > 255 {
		details["name"] = "name must be at most 255 characters"
	}
	if input.Year < 1900 || input.Year > 2100 {
		details["year"] = "year must be between 1900 and 2100"
	}
	return details
}

func extractID(r *http.Request) (int, error) {
	parts := strings.Split(r.URL.Path, "/")
	for i, p := range parts {
		if p == "series" && i+1 < len(parts) {
			return strconv.Atoi(parts[i+1])
		}
	}
	return 0, fmt.Errorf("no id in path")
}
