package reports

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/angellsan94/dogspots-api/internal/auth"
	"github.com/google/uuid"
)

type fakeRepository struct {
	placeID   uuid.UUID
	commentID uuid.UUID
	reason    string
	details   *string
	err       error
}

func (r *fakeRepository) CreateForPlace(_ context.Context, _ uuid.UUID, placeID uuid.UUID, reason string, details *string) error {
	r.placeID, r.reason, r.details = placeID, reason, details
	return r.err
}

func (r *fakeRepository) CreateForComment(_ context.Context, _ uuid.UUID, commentID uuid.UUID, reason string, details *string) error {
	r.commentID, r.reason, r.details = commentID, reason, details
	return r.err
}

func (*fakeRepository) List(context.Context, int) ([]ReportListItem, error) {
	return nil, nil
}

type handlerFunc func(*Handler) http.HandlerFunc

func placeReportHandler(h *Handler) http.HandlerFunc   { return h.CreatePlaceReport }
func commentReportHandler(h *Handler) http.HandlerFunc { return h.CreateCommentReport }

func callHandler(t *testing.T, repo Repository, handler handlerFunc, param, value, body string, authenticate bool) *httptest.ResponseRecorder {
	t.Helper()
	tokenManager := auth.NewTokenManager("test-secret", time.Minute)
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if authenticate {
		token, err := tokenManager.GenerateAccessToken(uuid.New())
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Authorization", "Bearer "+token)
	}
	request.SetPathValue(param, value)
	response := httptest.NewRecorder()
	auth.RequireAuth(tokenManager, handler(NewHandler(repo))).ServeHTTP(response, request)
	return response
}

func TestCreatePlaceReport(t *testing.T) {
	repo := &fakeRepository{}
	placeID := uuid.New()
	response := callHandler(t, repo, placeReportHandler, "placeId", placeID.String(), `{"reason":"dangerous","details":"  cristales  "}`, true)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if repo.placeID != placeID || repo.reason != ReasonDangerous || repo.details == nil || *repo.details != "cristales" {
		t.Fatalf("stored report = %#v", repo)
	}

	response = callHandler(t, repo, placeReportHandler, "placeId", placeID.String(), `{"reason":"dangerous"}`, true)
	if response.Code != http.StatusNoContent {
		t.Fatalf("duplicate report status = %d, body = %s", response.Code, response.Body.String())
	}
	if repo.details != nil {
		t.Fatalf("empty details should be nil, got %v", *repo.details)
	}
}

func TestCreateReportValidation(t *testing.T) {
	placeID := uuid.New()
	longDetails := strings.Repeat("x", 501)
	cases := []struct {
		name     string
		handler  handlerFunc
		param    string
		id       string
		body     string
		auth     bool
		repoErr  error
		wantCode int
	}{
		{name: "missing token", handler: placeReportHandler, param: "placeId", id: placeID.String(), body: `{"reason":"spam"}`, wantCode: http.StatusUnauthorized},
		{name: "invalid reason", handler: placeReportHandler, param: "placeId", id: placeID.String(), body: `{"reason":"boring"}`, auth: true, wantCode: http.StatusUnprocessableEntity},
		{name: "long details", handler: placeReportHandler, param: "placeId", id: placeID.String(), body: `{"reason":"other","details":"` + longDetails + `"}`, auth: true, wantCode: http.StatusUnprocessableEntity},
		{name: "unknown field", handler: placeReportHandler, param: "placeId", id: placeID.String(), body: `{"reason":"spam","extra":true}`, auth: true, wantCode: http.StatusUnprocessableEntity},
		{name: "invalid place id", handler: placeReportHandler, param: "placeId", id: "not-a-uuid", body: `{"reason":"spam"}`, auth: true, wantCode: http.StatusBadRequest},
		{name: "invalid comment id", handler: commentReportHandler, param: "commentId", id: "not-a-uuid", body: `{"reason":"spam"}`, auth: true, wantCode: http.StatusBadRequest},
		{name: "missing place", handler: placeReportHandler, param: "placeId", id: placeID.String(), body: `{"reason":"spam"}`, auth: true, repoErr: ErrPlaceNotFound, wantCode: http.StatusNotFound},
		{name: "missing comment", handler: commentReportHandler, param: "commentId", id: placeID.String(), body: `{"reason":"spam"}`, auth: true, repoErr: ErrCommentNotFound, wantCode: http.StatusNotFound},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			repo := &fakeRepository{err: test.repoErr}
			response := callHandler(t, repo, test.handler, test.param, test.id, test.body, test.auth)
			if response.Code != test.wantCode {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, test.wantCode, response.Body.String())
			}
		})
	}
}
