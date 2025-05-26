package response

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// JSON sends a JSON response with the given status code and data
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Interface("data", data).Msg("Failed to encode JSON response")
	}
}

// Success sends a successful JSON response with optional data
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// Created sends a 201 Created response with the created resource
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Paginated sends a paginated response with metadata
func Paginated(w http.ResponseWriter, data interface{}, page, pageSize, total int) {
	JSON(w, http.StatusOK, map[string]interface{}{
		"data": data,
		"pagination": map[string]int{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
			"pages":     (total + pageSize - 1) / pageSize,
		},
	})
}