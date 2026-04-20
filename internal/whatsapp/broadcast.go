package whatsapp

import (
	"context"
	"strings"
	"sync"
	"time"

	"wsapi/internal/domain"
)

type WorkerConfig struct {
	MaxWorkersPerEmpresa int
	MaxWorkersGlobal     int
	MaxRetries           int
	RetryDelay           time.Duration
}

var DefaultWorkerConfig = WorkerConfig{
	MaxWorkersPerEmpresa: 3,
	MaxWorkersGlobal:     20,
	MaxRetries:           2,
	RetryDelay:           1 * time.Second,
}

type BroadcastJob struct {
	ReferenceID string
	RUCEmpresa  string
	AccountID   string // NumeroCompleto del teléfono emisor (usado para SendTextMessage)
	Items       []domain.BroadcastItem
	ResultChan  chan BroadcastResult
}

type BroadcastResult struct {
	Index     int
	Destino   string
	State     string
	Error     string
	Timestamp time.Time
}

type BroadcastWorker struct {
	config    WorkerConfig
	manager   *Manager
	inputChan chan BroadcastJob

	empresaMu     sync.Mutex
	empresaQueues map[string]chan domain.BroadcastItem
	empresaCount  map[string]int

	activeWorkers   int
	activeWorkersMu sync.Mutex

	wg sync.WaitGroup
}

func NewBroadcastWorker(config WorkerConfig, manager *Manager) *BroadcastWorker {
	if config.MaxWorkersPerEmpresa == 0 {
		config.MaxWorkersPerEmpresa = DefaultWorkerConfig.MaxWorkersPerEmpresa
	}
	if config.MaxWorkersGlobal == 0 {
		config.MaxWorkersGlobal = DefaultWorkerConfig.MaxWorkersGlobal
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = DefaultWorkerConfig.MaxRetries
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = DefaultWorkerConfig.RetryDelay
	}

	return &BroadcastWorker{
		config:        config,
		manager:       manager,
		inputChan:     make(chan BroadcastJob, config.MaxWorkersGlobal*2),
		empresaQueues: make(map[string]chan domain.BroadcastItem),
		empresaCount:  make(map[string]int),
	}
}

func (w *BroadcastWorker) Start(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		w.wg.Add(1)
		go w.workerLoop()
	}
}

func (w *BroadcastWorker) workerLoop() {
	defer w.wg.Done()

	for {
		select {
		case job, ok := <-w.inputChan:
			if !ok {
				return
			}
			w.processJob(job)
		}
	}
}

func (w *BroadcastWorker) processJob(job BroadcastJob) {
	ruc := NormalizeAccountID(job.RUCEmpresa)

	w.ensureEmpresaQueue(ruc)

	for i, item := range job.Items {
		item := item
		result := BroadcastResult{
			Index:     i,
			Destino:   item.Destino,
			State:     "pending",
			Timestamp: time.Now(),
		}

		err := w.processItemWithRetry(item, job.AccountID)
		if err != nil {
			result.State = "failed"
			result.Error = err.Error()
		} else {
			result.State = "sent"
		}

		select {
		case job.ResultChan <- result:
		default:
		}
	}

	close(job.ResultChan)
	w.cleanupEmpresaQueue(ruc)
}

func (w *BroadcastWorker) processItemWithRetry(item domain.BroadcastItem, accountID string) error {
	var lastErr error

	for attempt := 0; attempt <= w.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(w.config.RetryDelay)
		}

		err := w.processItem(item, accountID)
		if err == nil {
			return nil
		}

		lastErr = err

		if !isTransientError(err) {
			return err
		}
	}

	return lastErr
}

func (w *BroadcastWorker) processItem(item domain.BroadcastItem, accountID string) error {
	if w.manager == nil {
		return ErrClientNotConnected
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return SendTextMessage(ctx, w.manager, accountID, item.Destino, item.Mensaje)
}

func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	transientMarkers := []string{"timeout", "connection", "reset", "refused", "temporary"}
	for _, marker := range transientMarkers {
		if strings.Contains(errStr, marker) {
			return true
		}
	}
	return false
}

func (w *BroadcastWorker) ensureEmpresaQueue(ruc string) {
	w.empresaMu.Lock()
	defer w.empresaMu.Unlock()

	if _, ok := w.empresaQueues[ruc]; !ok {
		w.empresaQueues[ruc] = make(chan domain.BroadcastItem, w.config.MaxWorkersPerEmpresa*2)
	}
	if w.empresaCount[ruc] >= w.config.MaxWorkersPerEmpresa {
		for w.empresaCount[ruc] >= w.config.MaxWorkersPerEmpresa {
			time.Sleep(100 * time.Millisecond)
		}
	}
	w.empresaCount[ruc]++
}

func (w *BroadcastWorker) cleanupEmpresaQueue(ruc string) {
	w.empresaMu.Lock()
	defer w.empresaMu.Unlock()

	if w.empresaCount[ruc] > 0 {
		w.empresaCount[ruc]--
	}
}

func (w *BroadcastWorker) Submit(job BroadcastJob) error {
	w.activeWorkersMu.Lock()
	if w.activeWorkers >= w.config.MaxWorkersGlobal {
		w.activeWorkersMu.Unlock()
		return ErrPoolFull
	}
	w.activeWorkers++
	w.activeWorkersMu.Unlock()

	w.inputChan <- job
	return nil
}

func (w *BroadcastWorker) SubmitAsync(job BroadcastJob) {
	go func() {
		if err := w.Submit(job); err != nil {
			close(job.ResultChan)
		}
	}()
}

func (w *BroadcastWorker) Shutdown() {
	close(w.inputChan)
	w.wg.Wait()
}

func (w *BroadcastWorker) Stats() (global int, perEmpresa map[string]int) {
	w.activeWorkersMu.Lock()
	global = w.activeWorkers
	w.activeWorkersMu.Unlock()

	w.empresaMu.Lock()
	perEmpresa = make(map[string]int, len(w.empresaCount))
	for k, v := range w.empresaCount {
		perEmpresa[k] = v
	}
	w.empresaMu.Unlock()

	return
}

var ErrPoolFull = &BroadcastError{Message: "worker pool is full"}

type BroadcastError struct {
	Message string
}

func (e *BroadcastError) Error() string {
	return e.Message
}


