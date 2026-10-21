package auth

import (
	"context"
	"log/slog"
	"net/http"

	apierrors "github.com/angellsan94/dogspots-api/internal/errors"
	"github.com/angellsan94/dogspots-api/internal/storage"
)

type AccountHandler struct {
	repo  *SQLiteRepository
	files storage.Storage
}

func NewAccountHandler(repo *SQLiteRepository, files storage.Storage) *AccountHandler {
	return &AccountHandler{repo: repo, files: files}
}

// DeleteMe elimina los datos de la cuenta y anonimiza el usuario.
func (h *AccountHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		apierrors.NewProblem(http.StatusUnauthorized, "Unauthorized", "Missing authenticated user", "unauthorized").Write(w)
		return
	}

	urls, err := h.repo.DeleteAccountData(r.Context(), userID)
	if err != nil {
		slog.Error("error al eliminar la cuenta", slog.Any("error", err), slog.String("user_id", userID.String()))
		apierrors.NewProblem(http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred", "internal_error").Write(w)
		return
	}

	// shortcut: si falla el borrado de un fichero queda huérfano en disco; añadir un limpiador periódico cuando haya almacenamiento externo.
	for _, url := range urls {
		if err := h.files.Delete(context.Background(), url); err != nil {
			slog.Warn("no se pudo borrar un fichero de una cuenta eliminada",
				slog.Any("error", err), slog.String("url", url), slog.String("user_id", userID.String()))
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
