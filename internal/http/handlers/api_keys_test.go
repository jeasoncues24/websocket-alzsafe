package http

import "testing"

func TestExtractAPIKeyID(t *testing.T) {
	if got := extractAPIKeyID("/api/admin/api-keys/123/rotate"); got != 123 {
		t.Fatalf("expected 123, got %d", got)
	}
}

func TestExtractTelefonoKeyID(t *testing.T) {
	if got := extractTelefonoKeyID("/api/admin/telefonos/77/api-keys", "/api/admin/telefonos/", "/api-keys"); got != 77 {
		t.Fatalf("expected 77, got %d", got)
	}
}
