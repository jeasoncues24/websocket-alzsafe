package http

import (
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/golang-jwt/jwt/v5"

	"wsapi/internal/config"
	"wsapi/internal/storage"
	"wsapi/internal/whatsapp"
)

func makeAdminToken(t *testing.T, secret string) string {
	t.Helper()
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  float64(1),
		"username": "root",
		"rol":      "super_admin",
		"is_root":  true,
		"exp":      time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("makeAdminToken: %v", err)
	}
	return tok
}

func readWSEvent(t *testing.T, ctx context.Context, conn *websocket.Conn) map[string]any {
	t.Helper()
	tctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, data, err := conn.Read(tctx)
	if err != nil {
		t.Fatalf("readWSEvent: %v", err)
	}
	var evt map[string]any
	if err := json.Unmarshal(data, &evt); err != nil {
		t.Fatalf("readWSEvent unmarshal: %v", err)
	}
	return evt
}

// TestConnectCompanyPhoneWS_AuthRequired verifica que sin token el servidor
// envía un evento "error" y cierra la conexión.
func TestConnectCompanyPhoneWS_AuthRequired(t *testing.T) {
	db := newAdminPhoneTestDB(t)
	insertAdminPhone(t, db, 1, "+51", "999888777", "+51999888777")

	h := &AdminHandler{
		telefonoStore: storage.NewTelefonoStore(db),
		sessionStore:  storage.NewSessionStore(),
		manager:       whatsapp.NewManager(),
		jwtCfg:        &config.JWTConfig{Secret: "test-secret", Issuer: "wsapi"},
	}

	srv := httptest.NewServer(stdhttp.HandlerFunc(h.ConnectCompanyPhoneWS))
	defer srv.Close()

	ctx := context.Background()
	wsURL := "ws" + srv.URL[4:] + "/api/admin/telefonos/1/connect/ws"

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	evt := readWSEvent(t, ctx, conn)
	if evt["event"] != "error" {
		t.Errorf("expected event=error, got %v", evt["event"])
	}
}

// TestConnectCompanyPhoneWS_TelefonoNotFound verifica que con token válido
// pero teléfono inexistente el servidor envía "error".
func TestConnectCompanyPhoneWS_TelefonoNotFound(t *testing.T) {
	db := newAdminPhoneTestDB(t)

	secret := "test-secret"
	h := &AdminHandler{
		telefonoStore: storage.NewTelefonoStore(db),
		sessionStore:  storage.NewSessionStore(),
		manager:       whatsapp.NewManager(),
		jwtCfg:        &config.JWTConfig{Secret: secret, Issuer: "wsapi"},
	}

	srv := httptest.NewServer(stdhttp.HandlerFunc(h.ConnectCompanyPhoneWS))
	defer srv.Close()

	token := makeAdminToken(t, secret)
	wsURL := "ws" + srv.URL[4:] + "/api/admin/telefonos/999/connect/ws?token=" + token

	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseNow()

	evt := readWSEvent(t, ctx, conn)
	if evt["event"] != "error" {
		t.Errorf("expected event=error, got %v", evt["event"])
	}
}

// TestConnectCompanyPhoneWS_QRSessionCleanedOnDisconnect verifica que cuando
// el WS se cierra mientras la sesión está en qr_pending, el manager ya no
// tiene registrado el accountID (el goroutine de sesión fue cancelado).
func TestConnectCompanyPhoneWS_QRSessionCleanedOnDisconnect(t *testing.T) {
	db := newAdminPhoneTestDB(t)
	insertAdminPhone(t, db, 1, "+51", "999888777", "+51999888777")

	secret := "test-secret"
	sessionStore := storage.NewSessionStore()
	manager := whatsapp.NewManager()

	// Simular sesión en qr_pending — como si el goroutine de sesión hubiera arrancado
	sessionStore.SetQRPending("+51999888777", "FAKE_QR_CODE")

	h := &AdminHandler{
		telefonoStore: storage.NewTelefonoStore(db),
		sessionStore:  sessionStore,
		manager:       manager,
		jwtCfg:        &config.JWTConfig{Secret: secret, Issuer: "wsapi"},
	}

	srv := httptest.NewServer(stdhttp.HandlerFunc(h.ConnectCompanyPhoneWS))
	defer srv.Close()

	token := makeAdminToken(t, secret)
	wsURL := "ws" + srv.URL[4:] + "/api/admin/telefonos/1/connect/ws?token=" + token

	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	// Leer el primer evento (phone-info)
	readWSEvent(t, ctx, conn)

	// Cerrar WS desde el cliente — simula cierre de tab
	conn.Close(websocket.StatusNormalClosure, "test done")

	// Esperar a que el manager limpie el accountID (con timeout de 1 segundo)
	accountID := whatsapp.NormalizeAccountID("+51999888777")
	cleaned := false
	for i := 0; i < 20; i++ {
		if !manager.Exists(accountID) {
			cleaned = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !cleaned {
		t.Errorf("expected manager to have cleaned up account %s after QR disconnect", accountID)
	}
}
