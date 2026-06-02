package whatsapp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeSQLiteFilename(t *testing.T) {
	tests := map[string]string{
		"20123456789":         "20123456789",
		" 20 12/34*56?789 ":   "20_12_34_56_789",
		"":                    "default",
		"@@@":                 "default",
		"telefono-empresa_01": "telefono-empresa_01",
	}

	for input, want := range tests {
		if got := sanitizeSQLiteFilename(input); got != want {
			t.Fatalf("sanitizeSQLiteFilename(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestStartSessionFallbackEmitsSnapshotEvents(t *testing.T) {
	manager := NewManager()

	events, unsubscribe, err := StartSession(manager, " 20123456789 ")
	if err != nil {
		t.Fatalf("StartSession returned error: %v", err)
	}
	defer unsubscribe()

	var got []SessionEvent
	for event := range events {
		got = append(got, event)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}

	if got[0].Event != "qr-20123456789" {
		t.Fatalf("expected first event qr-20123456789, got %s", got[0].Event)
	}
	if got[1].Event != "active-20123456789" {
		t.Fatalf("expected second event active-20123456789, got %s", got[1].Event)
	}
}

func TestIsWhatsmeowUpgradeConflictError(t *testing.T) {
	if !isWhatsmeowUpgradeConflictError(
		&testErr{"failed to run upgrade v0->v13: SQL logic error: table whatsmeow_device already exists (1)"},
	) {
		t.Fatal("expected upgrade conflict error to be ignored")
	}

	if isWhatsmeowUpgradeConflictError(&testErr{"some other sqlite error"}) {
		t.Fatal("unexpected match for unrelated error")
	}
}

func TestRemoveSQLiteArtifacts(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "phone.db")

	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := os.WriteFile(base+suffix, []byte("x"), 0o600); err != nil {
			t.Fatalf("write %s: %v", suffix, err)
		}
	}

	if err := removeSQLiteArtifacts(base); err != nil {
		t.Fatalf("removeSQLiteArtifacts: %v", err)
	}

	for _, suffix := range []string{"", "-wal", "-shm"} {
		if _, err := os.Stat(base + suffix); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed, got err=%v", suffix, err)
		}
	}
}

type testErr struct{ msg string }

func (e *testErr) Error() string { return e.msg }
