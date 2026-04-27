package auth

import "testing"

func TestGenerateAndParseAPIKeyMaterial(t *testing.T) {
	prefix, raw, err := GenerateAPIKeyMaterial()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if prefix == "" || raw == "" {
		t.Fatalf("expected non-empty prefix and raw key")
	}
	parsed, ok := ParseAPIKey(raw)
	if !ok {
		t.Fatalf("expected raw key to parse")
	}
	if parsed != prefix {
		t.Fatalf("expected parsed prefix %q, got %q", prefix, parsed)
	}
}

func TestParseAPIKeyRejectsInvalidFormat(t *testing.T) {
	if _, ok := ParseAPIKey("invalid-key"); ok {
		t.Fatalf("expected invalid format to be rejected")
	}
}
