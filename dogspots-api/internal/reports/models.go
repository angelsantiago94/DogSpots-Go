package reports

import (
	"time"

	"github.com/google/uuid"
)

const (
	ReasonSpam          = "spam"
	ReasonInappropriate = "inappropriate"
	ReasonWrongLocation = "wrong_location"
	ReasonDangerous     = "dangerous"
	ReasonOther         = "other"
)

func ValidReason(reason string) bool {
	switch reason {
	case ReasonSpam, ReasonInappropriate, ReasonWrongLocation, ReasonDangerous, ReasonOther:
		return true
	default:
		return false
	}
}

// Report es una denuncia registrada por un usuario sobre un lugar o un comentario.
type Report struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	PlaceID   *uuid.UUID
	CommentID *uuid.UUID
	Reason    string
	Details   *string
	CreatedAt time.Time
}

// ReportListItem añade los datos que necesita el CLI de moderación.
type ReportListItem struct {
	Report
	ReporterEmail string
	PlaceTitle    string
	CommentBody   string
}
