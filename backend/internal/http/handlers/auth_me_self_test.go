package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"wsapi/internal/domain"
	"wsapi/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

func TestUpdateMeAndMePassword(t *testing.T) {
	db := setupAuthModulesTestDB(t)
	userStore := storage.NewUserModuleStore(db) // Para reuse del setup
	_ = userStore

	// Crear usuarios directamente en BD
	user1Password := "secret123"
	hash1, _ := bcrypt.GenerateFromPassword([]byte(user1Password), bcrypt.DefaultCost)
	
	roleID := insertTestRole(t, db, "operator", false, `["companies"]`)

	res, err := db.Exec(`INSERT INTO admin_users (username, password_hash, email, role_id, activo) VALUES (?, ?, ?, ?, 1)`,
		"operator1", string(hash1), "operator1@test.com", roleID)
	if err != nil {
		t.Fatalf("crear operator1: %v", err)
	}
	user1ID, _ := res.LastInsertId()

	res, err = db.Exec(`INSERT INTO admin_users (username, password_hash, email, role_id, activo) VALUES (?, ?, ?, ?, 1)`,
		"operator2", string(hash1), "operator2@test.com", roleID)
	if err != nil {
		t.Fatalf("crear operator2: %v", err)
	}
	user2ID, _ := res.LastInsertId()
	_ = user2ID

	h := &AuthHandler{
		userStore: storage.NewAdminUserStore(db),
	}

	t.Run("UpdateMe_success", func(t *testing.T) {
		body := map[string]string{
			"username": "operator1_updated",
			"email":    "operator1_updated@test.com",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/auth/me", bytes.NewBuffer(jsonBody))
		
		// Inyectar JWT Claims
		claims := &domain.AdminJWTClaims{
			UserID:   user1ID,
			Username: "operator1",
			IsRoot:   false,
		}
		req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), claims))

		rr := httptest.NewRecorder()
		h.UpdateMe(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("esperaba status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		// Verificar que en BD se actualizó
		u, err := h.userStore.GetByID(user1ID)
		if err != nil {
			t.Fatalf("obtener usuario: %v", err)
		}
		if u.Username != "operator1_updated" || u.Email != "operator1_updated@test.com" {
			t.Errorf("datos no coinciden: username=%s, email=%s", u.Username, u.Email)
		}
	})

	t.Run("UpdateMe_conflict_email", func(t *testing.T) {
		body := map[string]string{
			"username": "operator1_conflict",
			"email":    "operator2@test.com", // El email de operator2
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/auth/me", bytes.NewBuffer(jsonBody))
		
		claims := &domain.AdminJWTClaims{
			UserID:   user1ID,
			Username: "operator1_updated",
			IsRoot:   false,
		}
		req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), claims))

		rr := httptest.NewRecorder()
		h.UpdateMe(rr, req)

		if rr.Code != http.StatusConflict {
			t.Errorf("esperaba status 409, got %d", rr.Code)
		}
	})

	t.Run("UpdateMePassword_success", func(t *testing.T) {
		body := map[string]string{
			"current_password": "secret123",
			"new_password":     "newsecret123",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/auth/me/password", bytes.NewBuffer(jsonBody))
		
		claims := &domain.AdminJWTClaims{
			UserID:   user1ID,
			Username: "operator1_updated",
			IsRoot:   false,
		}
		req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), claims))

		rr := httptest.NewRecorder()
		h.UpdateMePassword(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("esperaba status 200, got %d. Body: %s", rr.Code, rr.Body.String())
		}

		// Verificar que se puede verificar el nuevo password
		u, _ := h.userStore.GetByID(user1ID)
		err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("newsecret123"))
		if err != nil {
			t.Errorf("el nuevo password no se hasheó o guardó bien: %v", err)
		}
	})

	t.Run("UpdateMePassword_unauthorized_current_password", func(t *testing.T) {
		body := map[string]string{
			"current_password": "wrongpassword",
			"new_password":     "anothersecret",
		}
		jsonBody, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPut, "/api/auth/me/password", bytes.NewBuffer(jsonBody))
		
		claims := &domain.AdminJWTClaims{
			UserID:   user1ID,
			Username: "operator1_updated",
			IsRoot:   false,
		}
		req = req.WithContext(domain.WithAdminJWTClaims(req.Context(), claims))

		rr := httptest.NewRecorder()
		h.UpdateMePassword(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("esperaba status 401, got %d", rr.Code)
		}
	})
}

func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}
