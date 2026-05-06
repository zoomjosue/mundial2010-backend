package models

import "time"

type Serie struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Genre       string    `json:"genre"`
	Year        int       `json:"year"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SerieInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Genre       string `json:"genre"`
	Year        int    `json:"year"`
	ImageURL    string `json:"image_url"`
}

type Rating struct {
	ID        int       `json:"id"`
	SerieID   int       `json:"serie_id"`
	Score     int       `json:"score"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type RatingInput struct {
	Score   int    `json:"score"`
	Comment string `json:"comment"`
}

type RatingSummary struct {
	SerieID int      `json:"serie_id"`
	Average float64  `json:"average"`
	Count   int      `json:"count"`
	Ratings []Rating `json:"ratings"`
}

type PaginatedSeries struct {
	Data       []Serie `json:"data"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	Total      int     `json:"total"`
	TotalPages int     `json:"total_pages"`
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
