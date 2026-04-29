package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/example/job-crud-app/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Jobs struct {
	DB  *pgxpool.Pool
	Log *slog.Logger
}

func (h Jobs) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(r.Context(), `
		SELECT id::text, title, company, description, status, created_at, updated_at
		FROM jobs
		ORDER BY created_at DESC
	`)
	if err != nil {
		h.Log.ErrorContext(r.Context(), "list jobs query", slog.Any("err", err))
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []models.Job
	for rows.Next() {
		var j models.Job
		if err := rows.Scan(&j.ID, &j.Title, &j.Company, &j.Description, &j.Status, &j.CreatedAt, &j.UpdatedAt); err != nil {
			h.Log.ErrorContext(r.Context(), "list jobs scan", slog.Any("err", err))
			http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
			return
		}
		list = append(list, j)
	}
	if list == nil {
		list = []models.Job{}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		h.Log.ErrorContext(r.Context(), "encode job list", slog.Any("err", err))
	}
}

func (h Jobs) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var j models.Job
	err := h.DB.QueryRow(r.Context(), `
		SELECT id::text, title, company, description, status, created_at, updated_at
		FROM jobs WHERE id = $1
	`, id).Scan(&j.ID, &j.Title, &j.Company, &j.Description, &j.Status, &j.CreatedAt, &j.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, `{"error":"Job not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		h.Log.ErrorContext(r.Context(), "get job", slog.String("id", id), slog.Any("err", err))
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(j); err != nil {
		h.Log.ErrorContext(r.Context(), "encode job", slog.Any("err", err))
	}
}

type jobBody struct {
	Title       string  `json:"title"`
	Company     *string `json:"company"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
}

func (h Jobs) Create(w http.ResponseWriter, r *http.Request) {
	var body jobBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
		return
	}
	status := "open"
	if body.Status != nil && strings.TrimSpace(*body.Status) != "" {
		status = strings.TrimSpace(*body.Status)
	}
	var company interface{}
	if body.Company != nil {
		s := strings.TrimSpace(*body.Company)
		if s != "" {
			company = s
		}
	}
	var desc interface{}
	if body.Description != nil {
		desc = *body.Description
	}

	var j models.Job
	err := h.DB.QueryRow(r.Context(), `
		INSERT INTO jobs (title, company, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, title, company, description, status, created_at, updated_at
	`, title, company, desc, status).Scan(
		&j.ID, &j.Title, &j.Company, &j.Description, &j.Status, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		h.Log.ErrorContext(r.Context(), "create job", slog.Any("err", err))
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(j); err != nil {
		h.Log.ErrorContext(r.Context(), "encode created job", slog.Any("err", err))
	}
}

func (h Jobs) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tag, err := h.DB.Exec(r.Context(), `DELETE FROM jobs WHERE id = $1`, id)
	if err != nil {
		h.Log.ErrorContext(r.Context(), "delete job", slog.String("id", id), slog.Any("err", err))
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, `{"error":"Job not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
