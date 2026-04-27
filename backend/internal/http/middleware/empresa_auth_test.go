package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"wsapi/internal/auth"
	"wsapi/internal/config"
	"wsapi/internal/domain"
)

type fakeEmpresaStore struct {
	empresa *domain.Empresa
}

func (f fakeEmpresaStore) GetByID(id int64) (*domain.Empresa, error) {
	if f.empresa != nil && f.empresa.ID == id {
		return f.empresa, nil
	}
	return nil, nil
}

func (fakeEmpresaStore) GetByRUC(string) (*domain.Empresa, error) { return nil, nil }
func (fakeEmpresaStore) GetAll(int, int, string, *bool) ([]domain.Empresa, int, error) {
	return nil, 0, nil
}
func (fakeEmpresaStore) Create(*domain.Empresa) (int64, error) { return 0, nil }
func (fakeEmpresaStore) Update(*domain.Empresa) error          { return nil }
func (fakeEmpresaStore) Delete(int64) error                    { return nil }
func (fakeEmpresaStore) IncrementTokenVersion(int64) (int, error) {
	return 0, nil
}

func TestRequireEmpresaAuthAcceptsEmpresaJWT(t *testing.T) {
	secret := "test-secret"
	store := fakeEmpresaStore{empresa: &domain.Empresa{ID: 1, RUC: "20123456789", Nombre: "Acme", TokenVersion: 1, Activo: true}}
	mw := NewEmpresaAuthMiddleware(&config.JWTConfig{Secret: secret}, store, nil)

	token, err := auth.GenerateEmpresaJWT(store.empresa, secret, "wsapi")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/modules", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	mw.RequireEmpresaAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := domain.GetEmpresaJWTClaims(r.Context())
		if !ok || claims == nil || claims.EmpresaID != 1 {
			t.Fatalf("expected empresa claims in context, got %+v", claims)
		}
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRequireEmpresaAuthRejectsMissingOrAdminJWT(t *testing.T) {
	secret := "test-secret"
	store := fakeEmpresaStore{empresa: &domain.Empresa{ID: 1, RUC: "20123456789", Nombre: "Acme", TokenVersion: 1, Activo: true}}
	mw := NewEmpresaAuthMiddleware(&config.JWTConfig{Secret: secret}, store, nil)
	wrapped := mw.RequireEmpresaAuth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("missing auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/modules", nil)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})

	t.Run("admin jwt", func(t *testing.T) {
		adminToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  99,
			"username": "root",
			"rol":      "super_admin",
			"is_root":  true,
		})
		signed, err := adminToken.SignedString([]byte(secret))
		if err != nil {
			t.Fatalf("sign token: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/admin/modules", nil)
		req.Header.Set("Authorization", "Bearer "+signed)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d body=%s", rr.Code, rr.Body.String())
		}
	})
}
