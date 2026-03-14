package oidc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/Servora-Kit/servora/app/iam/service/internal/biz"
	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
	dataent "github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"
)

type fakeAuthnRepo struct {
	users map[string]*entity.User
}

func newFakeAuthnRepo() *fakeAuthnRepo {
	return &fakeAuthnRepo{users: make(map[string]*entity.User)}
}

func (r *fakeAuthnRepo) SaveUser(_ context.Context, u *entity.User) (*entity.User, error) {
	r.users[u.Email] = u
	return u, nil
}

func (r *fakeAuthnRepo) GetUserByEmail(_ context.Context, email string) (*entity.User, error) {
	u, ok := r.users[email]
	if !ok {
		return nil, &dataent.NotFoundError{}
	}
	return u, nil
}

func (r *fakeAuthnRepo) GetUserByUserName(context.Context, string) (*entity.User, error) {
	return nil, &dataent.NotFoundError{}
}

func (r *fakeAuthnRepo) GetUserByID(context.Context, string) (*entity.User, error) {
	return nil, &dataent.NotFoundError{}
}

func (r *fakeAuthnRepo) UpdatePassword(context.Context, string, string) error { return nil }

func (r *fakeAuthnRepo) UpdateEmailVerified(context.Context, string, bool) error { return nil }

func (r *fakeAuthnRepo) SaveRefreshToken(_ context.Context, _ string, _ string, _ time.Duration) error {
	return nil
}

func (r *fakeAuthnRepo) GetRefreshToken(context.Context, string) (string, error) {
	return "", nil
}

func (r *fakeAuthnRepo) DeleteRefreshToken(context.Context, string) error { return nil }

func (r *fakeAuthnRepo) DeleteUserRefreshTokens(context.Context, string) error { return nil }

var _ biz.AuthnRepo = (*fakeAuthnRepo)(nil)

func newTestLoginHandler(repo biz.AuthnRepo) *LoginHandler {
	return NewLoginHandler(repo, nil, log.DefaultLogger)
}

func TestLoginHandler_GET_MissingAuthRequestID(t *testing.T) {
	h := newTestLoginHandler(newFakeAuthnRepo())

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "missing authRequestID") {
		t.Fatalf("expected body to contain 'missing authRequestID', got %q", rec.Body.String())
	}
}

func TestLoginHandler_GET_RenderForm(t *testing.T) {
	h := newTestLoginHandler(newFakeAuthnRepo())

	req := httptest.NewRequest(http.MethodGet, "/login?authRequestID=test-123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "test-123") {
		t.Fatalf("expected rendered form to contain authRequestID, got %q", body)
	}
	if !strings.Contains(body, "<form") {
		t.Fatalf("expected rendered form to contain <form> tag, got %q", body)
	}
}

func TestLoginCompleteHandler_BadJSON(t *testing.T) {
	lh := newTestLoginHandler(newFakeAuthnRepo())
	h := NewLoginCompleteHandler(lh)

	req := httptest.NewRequest(http.MethodPost, "/login/complete", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestLoginHandler_MethodNotAllowed(t *testing.T) {
	h := newTestLoginHandler(newFakeAuthnRepo())

	req := httptest.NewRequest(http.MethodDelete, "/login", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestLoginCompleteHandler_MethodNotAllowed(t *testing.T) {
	lh := newTestLoginHandler(newFakeAuthnRepo())
	h := NewLoginCompleteHandler(lh)

	req := httptest.NewRequest(http.MethodGet, "/login/complete", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestLoginCompleteHandler_MissingFields(t *testing.T) {
	t.Skip("requires Redis for full authenticate flow")
}

func TestLoginHandler_POST_AuthenticateFlow(t *testing.T) {
	t.Skip("requires Redis for full authenticate flow")
}
