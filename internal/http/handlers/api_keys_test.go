package http

import (
	"testing"

	"wsapi/internal/domain"
)

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

func TestBuildSessionRuntimeInfoMismatch(t *testing.T) {
	phone := &domain.Telefono{ID: 5, Status: domain.TelefonoStatusActive}
	info := buildSessionRuntimeInfo(phone, "51912345678", false)

	if info["mismatch"] != true {
		t.Fatalf("expected mismatch true, got %v", info["mismatch"])
	}
	if info["mismatch_reason"] != "db_active_runtime_disconnected" {
		t.Fatalf("unexpected mismatch reason: %v", info["mismatch_reason"])
	}
	if info["recommended_action"] != "reanudar_conexion" {
		t.Fatalf("unexpected action: %v", info["recommended_action"])
	}
}

func TestBuildSessionRuntimeInfoHealthy(t *testing.T) {
	phone := &domain.Telefono{ID: 7, Status: domain.TelefonoStatusActive}
	info := buildSessionRuntimeInfo(phone, "51912345678", true)

	if info["mismatch"] != false {
		t.Fatalf("expected mismatch false, got %v", info["mismatch"])
	}
	if info["status_runtime"] != "connected" {
		t.Fatalf("expected connected runtime status, got %v", info["status_runtime"])
	}
	if info["recommended_action"] != "none" {
		t.Fatalf("unexpected action: %v", info["recommended_action"])
	}
}
