package errors

import (
	"encoding/json"
	"net/http"
)

// Problem representa un error siguiendo el estándar RFC 7807 (Problem Details for HTTP APIs).
type Problem struct {
	Type     string `json:"type"`     // URI que identifica el tipo de problema
	Title    string `json:"title"`    // Breve resumen del problema
	Status   int    `json:"status"`   // Código de estado HTTP
	Detail   string `json:"detail,omitempty"` // Explicación detallada
	Instance string `json:"instance,omitempty"` // URI que identifica la ocurrencia del problema
	Code     string `json:"code"`     // Código de error interno y estable para el cliente
}

// NewProblem crea una nueva instancia de Problem.
func NewProblem(status int, title, detail, code string) *Problem {
	return &Problem{
		Type:   "about:blank", // Por defecto, un tipo genérico
		Title:  title,
		Status: status,
		Detail: detail,
		Code:   code,
	}
}

// Write escribe la respuesta de error en el http.ResponseWriter.
func (p *Problem) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}

// Common error helpers
var (
	ErrInternal = NewProblem(http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred on the server", "internal_error")
	ErrNotFound = NewProblem(http.StatusNotFound, "Not Found", "The requested resource was not found", "not_found")
	ErrBadRequest = NewProblem(http.StatusBadRequest, "Bad Request", "The request is invalid", "bad_request")
)
