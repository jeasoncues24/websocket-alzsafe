package storage

import (
	"sync"
	"time"

	"wsapi/internal/domain"
)

type BroadcastStore struct {
	mu   sync.RWMutex
	jobs map[string]*domain.BroadcastJob
}

func NewBroadcastStore() *BroadcastStore {
	return &BroadcastStore{
		jobs: make(map[string]*domain.BroadcastJob),
	}
}

func (s *BroadcastStore) Create(job *domain.BroadcastJob) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	job.Status = domain.BroadcastStatusPending
	job.Results = make([]domain.BroadcastResult, 0, job.Total)

	s.jobs[job.ReferenceID] = job
}

func (s *BroadcastStore) Get(referenceID string) (*domain.BroadcastJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[referenceID]
	return job, ok
}

func (s *BroadcastStore) AppendResult(referenceID string, result domain.BroadcastResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[referenceID]
	if !ok {
		return
	}

	job.Results = append(job.Results, result)
	job.UpdatedAt = time.Now()
}

func (s *BroadcastStore) UpdateStatus(referenceID string, status domain.BroadcastStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[referenceID]
	if !ok {
		return
	}

	job.Status = status
	job.UpdatedAt = time.Now()
}

func (s *BroadcastStore) ListByRUC(ruc string) []*domain.BroadcastJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var jobs []*domain.BroadcastJob
	for _, job := range s.jobs {
		if job.RUCEmpresa == ruc {
			jobs = append(jobs, job)
		}
	}

	return jobs
}
