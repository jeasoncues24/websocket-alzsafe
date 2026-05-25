package storage

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"wsapi/internal/domain"
)

func newJobQueueTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	_, err = db.Exec(`
CREATE TABLE job_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 5,
    empresa_id INTEGER NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    last_heartbeat TIMESTAMP NULL,
    next_retry_at TIMESTAMP NULL,
    metadata TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL
);
CREATE TABLE job_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id INTEGER NOT NULL,
    sequence_order INTEGER NOT NULL,
    payload TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    attempt_count INTEGER NOT NULL DEFAULT 0,
    error_text TEXT NULL,
    processed_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(job_id) REFERENCES job_queue(id) ON DELETE CASCADE
);`)
	if err != nil {
		_ = db.Close()
		t.Fatalf("create schema: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestCreateJobWithItems(t *testing.T) {
	db := newJobQueueTestDB(t)
	repo := NewJobQueueRepository(db)
	ctx := context.Background()

	job := &domain.Job{
		Type:        domain.JobTypeBroadcast,
		EntityID:    "test-ref-id",
		Priority:    5,
		EmpresaID:   123,
		MaxAttempts: 3,
		Metadata:    `{"test":true}`,
	}

	items := []domain.JobItem{
		{SequenceOrder: 0, Payload: `{"destino":"123","mensaje":"hi"}`},
		{SequenceOrder: 1, Payload: `{"destino":"456","mensaje":"hola"}`},
	}

	err := repo.CreateJobWithItems(ctx, job, items)
	if err != nil {
		t.Fatalf("CreateJobWithItems failed: %v", err)
	}

	if job.ID == 0 {
		t.Error("expected job ID to be set, got 0")
	}

	// Recuperar job y verificar
	savedJob, err := repo.GetByEntityID(ctx, "test-ref-id")
	if err != nil {
		t.Fatalf("GetByEntityID failed: %v", err)
	}
	if savedJob == nil {
		t.Fatal("expected to find job, got nil")
	}
	if savedJob.Metadata != `{"test":true}` {
		t.Errorf("expected metadata to be %q, got %q", `{"test":true}`, savedJob.Metadata)
	}

	// Recuperar items
	savedItems, err := repo.GetAllItems(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetAllItems failed: %v", err)
	}
	if len(savedItems) != 2 {
		t.Errorf("expected 2 items, got %d", len(savedItems))
	}
}

func TestRecoverStuckJobs(t *testing.T) {
	db := newJobQueueTestDB(t)
	repo := NewJobQueueRepository(db)
	ctx := context.Background()

	// Crear un job corriendo
	job := &domain.Job{
		Type:        domain.JobTypeBroadcast,
		EntityID:    "stuck-job-id",
		Priority:    5,
		EmpresaID:   123,
		MaxAttempts: 3,
	}
	err := repo.CreateJobWithItems(ctx, job, nil)
	if err != nil {
		t.Fatalf("CreateJobWithItems failed: %v", err)
	}

	err = repo.UpdateStatus(ctx, job.ID, domain.JobStatusRunning, nil)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	// Simular que el heartbeat está atascado poniéndole una fecha antigua en el pasado
	_, err = db.Exec("UPDATE job_queue SET last_heartbeat = ? WHERE id = ?", time.Now().Add(-10*time.Minute), job.ID)
	if err != nil {
		t.Fatalf("update last_heartbeat: %v", err)
	}

	// Ejecutar recuperación
	recovered, err := repo.RecoverStuckJobs(ctx, 5*time.Minute)
	if err != nil {
		t.Fatalf("RecoverStuckJobs failed: %v", err)
	}
	if recovered != 1 {
		t.Errorf("expected 1 job recovered, got %d", recovered)
	}

	// Verificar que el job volvió a pending
	savedJob, err := repo.GetByEntityID(ctx, "stuck-job-id")
	if err != nil {
		t.Fatalf("GetByEntityID failed: %v", err)
	}
	if savedJob.Status != domain.JobStatusPending {
		t.Errorf("expected status pending, got %s", savedJob.Status)
	}
	if savedJob.AttemptCount != 1 {
		t.Errorf("expected attempt count to be 1, got %d", savedJob.AttemptCount)
	}
}

func TestUpdateItemStatus(t *testing.T) {
	db := newJobQueueTestDB(t)
	repo := NewJobQueueRepository(db)
	ctx := context.Background()

	job := &domain.Job{
		Type:        domain.JobTypeBroadcast,
		EntityID:    "job-id",
		Priority:    5,
		EmpresaID:   123,
		MaxAttempts: 3,
	}
	items := []domain.JobItem{
		{SequenceOrder: 0, Payload: `{"destino":"123","mensaje":"hi"}`},
	}
	err := repo.CreateJobWithItems(ctx, job, items)
	if err != nil {
		t.Fatalf("CreateJobWithItems failed: %v", err)
	}

	savedItems, err := repo.GetAllItems(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetAllItems failed: %v", err)
	}

	itemID := savedItems[0].ID
	err = repo.UpdateItemStatus(ctx, itemID, domain.JobItemSent, "no error")
	if err != nil {
		t.Fatalf("UpdateItemStatus failed: %v", err)
	}

	// Verificar
	updatedItems, err := repo.GetAllItems(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetAllItems failed: %v", err)
	}
	if updatedItems[0].Status != domain.JobItemSent {
		t.Errorf("expected status sent, got %s", updatedItems[0].Status)
	}
	if updatedItems[0].ErrorText != "no error" {
		t.Errorf("expected error text 'no error', got %q", updatedItems[0].ErrorText)
	}
}
