package storage

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedMigrationsMatchNormalizedLayout(t *testing.T) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	got := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		got[entry.Name()] = true
	}

	expected := []string{
		"000001_create_messages_table.up.sql",
		"000001_create_messages_table.down.sql",
		"000002_create_broadcasts_table.up.sql",
		"000002_create_broadcasts_table.down.sql",
		"000003_create_broadcast_results_table.up.sql",
		"000003_create_broadcast_results_table.down.sql",
		"000004_create_admin_users_table.up.sql",
		"000004_create_admin_users_table.down.sql",
		"000005_create_empresas_table.up.sql",
		"000005_create_empresas_table.down.sql",
		"000006_create_telefonos_table.up.sql",
		"000006_create_telefonos_table.down.sql",
		"000007_create_roles_table.up.sql",
		"000007_create_roles_table.down.sql",
		"000008_create_modules_table.up.sql",
		"000008_create_modules_table.down.sql",
		"000009_create_user_modules_table.up.sql",
		"000009_create_user_modules_table.down.sql",
		"000010_create_token_blacklist_table.up.sql",
		"000010_create_token_blacklist_table.down.sql",
		"000011_create_api_keys_table.up.sql",
		"000011_create_api_keys_table.down.sql",
		"000012_create_telefono_request_logs.up.sql",
		"000012_create_telefono_request_logs.down.sql",
		"000013_create_telefono_metrics_min.up.sql",
		"000013_create_telefono_metrics_min.down.sql",
		"000014_create_api_key_audit_events_table.up.sql",
		"000014_create_api_key_audit_events_table.down.sql",
		"000015_create_audit_log_table.up.sql",
		"000015_create_audit_log_table.down.sql",
		"000016_seeds.up.sql",
		"000016_seeds.down.sql",
		"000017_create_webhooks_outbound.up.sql",
		"000017_create_webhooks_outbound.down.sql",
		"000018_create_webhooks_outbound_queue.up.sql",
		"000018_create_webhooks_outbound_queue.down.sql",
		"000019_create_job_queue.up.sql",
		"000019_create_job_queue.down.sql",
	}

	if len(got) != len(expected) {
		t.Fatalf("expected %d migration files, got %d", len(expected), len(got))
	}

	for _, name := range expected {
		if !got[name] {
			t.Fatalf("expected embedded migration %s", name)
		}
	}

	unexpected := []string{
		"000012_create_api_key_usage_events_table.up.sql",
		"000012_create_api_key_usage_events_table.down.sql",
		"000013_create_api_key_usage_daily_table.up.sql",
		"000013_create_api_key_usage_daily_table.down.sql",
		"000017_create_telemetry_tables.up.sql",
		"000017_create_telemetry_tables.down.sql",
		"000015_add_missing_columns.up.sql",
		"000015_add_missing_columns.down.sql",
		"000016_rename_empresa_telefono_contacto.up.sql",
		"000016_rename_empresa_telefono_contacto.down.sql",
		"000017_align_messages_schema_with_repository.up.sql",
		"000017_align_messages_schema_with_repository.down.sql",
		"000018_add_retry_fields_to_messages.up.sql",
		"000018_add_retry_fields_to_messages.down.sql",
		"000019_create_audit_log_table.up.sql",
		"000019_create_audit_log_table.down.sql",
		"000020_add_is_root_to_roles.up.sql",
		"000020_add_is_root_to_roles.down.sql",
		"000022_add_audit_columns_to_empresas.up.sql",
		"000022_add_audit_columns_to_empresas.down.sql",
		"000023_add_audit_columns_to_telefonos.up.sql",
		"000023_add_audit_columns_to_telefonos.down.sql",
		"000024_add_audit_columns_to_roles.up.sql",
		"000024_add_audit_columns_to_roles.down.sql",
		"000025_add_audit_columns_to_api_keys.up.sql",
		"000025_add_audit_columns_to_api_keys.down.sql",
	}

	for _, name := range unexpected {
		if got[name] {
			t.Fatalf("unexpected legacy migration still embedded: %s", name)
		}
	}
}

func TestNormalizedCreateTableMigrationsHaveFinalSchema(t *testing.T) {
	cases := []struct {
		name             string
		requiredContains []string
		forbidden        []string
		createCount      int
	}{
		{
			name: "000001_create_messages_table.up.sql",
			requiredContains: []string{"adjuntos_json", "error_reason", "retry_count", "last_attempt_at", "timestamp_created", "timestamp_sent", "timestamp_confirmed"},
			forbidden:        []string{"ALTER TABLE", "INSERT INTO"},
			createCount:      1,
		},
		{
			name:             "000004_create_admin_users_table.up.sql",
			forbidden:        []string{"INSERT INTO"},
			createCount:      1,
		},
		{
			name:             "000005_create_empresas_table.up.sql",
			requiredContains: []string{"telefono_contacto VARCHAR(30)"},
			forbidden:        []string{"telefono VARCHAR(30)", "INSERT INTO", "ALTER TABLE"},
			createCount:      1,
		},
		{
			name:             "000007_create_roles_table.up.sql",
			forbidden:        []string{"INSERT INTO"},
			createCount:      1,
		},
		{
			name:             "000008_create_modules_table.up.sql",
			requiredContains: []string{"slug VARCHAR(50) UNIQUE"},
			forbidden:        []string{"INSERT INTO", "ALTER TABLE"},
			createCount:      1,
		},
		{
			name:             "000009_create_user_modules_table.up.sql",
			forbidden:        []string{"INSERT INTO"},
			createCount:      1,
		},
		{
			name:             "000011_create_api_keys_table.up.sql",
			requiredContains: []string{"updated_by BIGINT NULL", "INDEX idx_api_keys_updated_by (updated_by)"},
			forbidden:        []string{"ALTER TABLE"},
			createCount:      1,
		},
		{
			name:             "000015_create_audit_log_table.up.sql",
			requiredContains: []string{"CREATE TABLE IF NOT EXISTS audit_log"},
			forbidden:        []string{"ALTER TABLE", "INSERT INTO"},
			createCount:      1,
		},
		{
			name:             "000017_create_webhooks_outbound.up.sql",
			requiredContains: []string{"webhooks_outbound", "empresa_id", "telefono_id", "api_key_id", "url", "secret", "eventos", "idx_webhooks_empresa", "idx_webhooks_empresa_created", "idx_webhooks_telefono", "idx_webhooks_api_key"},
			forbidden:        []string{"ALTER TABLE", "INSERT INTO"},
			createCount:      1,
		},
		{
			name:             "000018_create_webhooks_outbound_queue.up.sql",
			requiredContains: []string{"webhooks_outbound_queue", "webhook_id", "payload", "estado", "idx_queue_due"},
			forbidden:        []string{"ALTER TABLE", "INSERT INTO"},
			createCount:      1,
		},
		{
			name:             "000019_create_job_queue.up.sql",
			requiredContains: []string{"job_queue", "job_items", "entity_id", "empresa_id", "idx_empresa_status", "idx_job_status_seq"},
			forbidden:        []string{"ALTER TABLE", "INSERT INTO"},
			createCount:      2,
		},
		{
			name:             "000012_create_telefono_request_logs.up.sql",
			requiredContains: []string{"telefono_request_logs", "api_key_id", "contract_name", "latency_ms", "idx_trl_key_time"},
			forbidden:        []string{"ALTER TABLE", "api_key_usage_events"},
			createCount:      0,
		},
		{
			name:             "000013_create_telefono_metrics_min.up.sql",
			requiredContains: []string{"telefono_metrics_min", "bucket_min", "latency_p50_ms", "uq_tmm_bucket"},
			forbidden:        []string{"ALTER TABLE", "api_key_usage_daily"},
			createCount:      0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			content := mustReadMigration(t, tc.name)
			if got := strings.Count(content, "CREATE TABLE IF NOT EXISTS"); got != tc.createCount {
				t.Fatalf("expected %d CREATE TABLE statements, got %d", tc.createCount, got)
			}
			for _, want := range tc.requiredContains {
				if !strings.Contains(content, want) {
					t.Fatalf("expected %s to contain %q", tc.name, want)
				}
			}
			for _, forbidden := range tc.forbidden {
				if strings.Contains(content, forbidden) {
					t.Fatalf("expected %s not to contain %q", tc.name, forbidden)
				}
			}
		})
	}
}

func TestSeedsMigrationContainsAllInitialDataAndReversibleDeletes(t *testing.T) {
	up := mustReadMigration(t, "000016_seeds.up.sql")
	for _, want := range []string{
		"INSERT IGNORE INTO roles",
		"INSERT IGNORE INTO admin_users",
		"INSERT IGNORE INTO modules",
		"INSERT IGNORE INTO user_modules",
	} {
		if !strings.Contains(up, want) {
			t.Fatalf("expected seeds up migration to contain %q", want)
		}
	}

	down := mustReadMigration(t, "000016_seeds.down.sql")
	positions := []struct {
		name string
		idx  int
	}{
		{"user_modules", strings.Index(down, "DELETE FROM user_modules")},
		{"admin_users", strings.Index(down, "DELETE FROM admin_users")},
		{"modules", strings.Index(down, "DELETE FROM modules")},
		{"roles", strings.Index(down, "DELETE FROM roles")},
	}
	for _, pos := range positions {
		if pos.idx == -1 {
			t.Fatalf("expected down migration to delete from %s", pos.name)
		}
	}
	if !(positions[0].idx < positions[1].idx && positions[1].idx < positions[2].idx && positions[2].idx < positions[3].idx) {
		t.Fatalf("expected delete order to be user_modules -> admin_users -> modules -> roles")
	}
}

func mustReadMigration(t *testing.T, name string) string {
	t.Helper()
	content, err := migrationsFS.ReadFile("migrations/" + name)
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	return string(content)
}
