package whatsapp

import (
	"context"
	"math/rand"
	"strings"
	"sync"
	"time"

	wa "go.mau.fi/whatsmeow"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

// WorkerConfig configura el pool de workers y los delays anti-ban.
type WorkerConfig struct {
	MaxWorkersPerEmpresa int
	MaxWorkersGlobal     int
	MaxRetries           int
	RetryDelay           time.Duration

	// Batching
	BatchSizeMin int
	BatchSizeMax int

	// Delay entre mensajes dentro del mismo batch
	IntraBatchDelayMin time.Duration
	IntraBatchDelayMax time.Duration

	// Delay entre batches
	InterBatchDelayMin time.Duration
	InterBatchDelayMax time.Duration

	// Macro-pausa cada N mensajes enviados
	MacroPauseEvery int
	MacroPauseMin   time.Duration
	MacroPauseMax   time.Duration
}

var DefaultWorkerConfig = WorkerConfig{
	MaxWorkersPerEmpresa: 3,
	MaxWorkersGlobal:     20,
	MaxRetries:           2,
	RetryDelay:           2 * time.Second,

	BatchSizeMin: 3,
	BatchSizeMax: 4,

	IntraBatchDelayMin: 1500 * time.Millisecond,
	IntraBatchDelayMax: 4000 * time.Millisecond,

	InterBatchDelayMin: 3000 * time.Millisecond,
	InterBatchDelayMax: 8000 * time.Millisecond,

	MacroPauseEvery: 10,
	MacroPauseMin:   15 * time.Second,
	MacroPauseMax:   30 * time.Second,
}

// TimingConfig retorna la configuración de timing del worker para calcular estimados.
func (c WorkerConfig) TimingConfig() domain.BroadcastTimingConfig {
	return domain.BroadcastTimingConfig{
		BatchSizeMin:       c.BatchSizeMin,
		BatchSizeMax:       c.BatchSizeMax,
		IntraBatchDelayMin: c.IntraBatchDelayMin,
		IntraBatchDelayMax: c.IntraBatchDelayMax,
		InterBatchDelayMin: c.InterBatchDelayMin,
		InterBatchDelayMax: c.InterBatchDelayMax,
		MacroPauseEvery:    c.MacroPauseEvery,
		MacroPauseMin:      c.MacroPauseMin,
		MacroPauseMax:      c.MacroPauseMax,
	}
}

// BroadcastJob describe un trabajo de difusión a procesar.
type BroadcastJob struct {
	ReferenceID string
	RUCEmpresa  string
	AccountID   string
	JobID       int64 // ID en job_queue (0 si no hay persistencia)
	Attachments []domain.AttachmentPayload
	Items       []domain.BroadcastItem
	ItemIDs     []int64 // IDs de los job_items en DB (paralelo a Items)
}

// BroadcastResult resultado de un item individual.
type BroadcastResult struct {
	Index     int
	Destino   string
	State     string
	Error     string
	Timestamp time.Time
}

// BroadcastWorker procesa jobs de difusión de forma asíncrona con delays anti-ban.
type BroadcastWorker struct {
	config    WorkerConfig
	manager   *Manager
	repo      storage.JobQueueRepository // nil si no hay DB
	inputChan chan BroadcastJob

	empresaMu    sync.Mutex
	empresaCount map[string]int

	activeWorkers   int
	activeWorkersMu sync.Mutex
	stopped         bool

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
	if config.BatchSizeMin == 0 {
		config.BatchSizeMin = DefaultWorkerConfig.BatchSizeMin
	}
	if config.BatchSizeMax == 0 {
		config.BatchSizeMax = DefaultWorkerConfig.BatchSizeMax
	}
	if config.IntraBatchDelayMin == 0 {
		config.IntraBatchDelayMin = DefaultWorkerConfig.IntraBatchDelayMin
	}
	if config.IntraBatchDelayMax == 0 {
		config.IntraBatchDelayMax = DefaultWorkerConfig.IntraBatchDelayMax
	}
	if config.InterBatchDelayMin == 0 {
		config.InterBatchDelayMin = DefaultWorkerConfig.InterBatchDelayMin
	}
	if config.InterBatchDelayMax == 0 {
		config.InterBatchDelayMax = DefaultWorkerConfig.InterBatchDelayMax
	}
	if config.MacroPauseEvery == 0 {
		config.MacroPauseEvery = DefaultWorkerConfig.MacroPauseEvery
	}
	if config.MacroPauseMin == 0 {
		config.MacroPauseMin = DefaultWorkerConfig.MacroPauseMin
	}
	if config.MacroPauseMax == 0 {
		config.MacroPauseMax = DefaultWorkerConfig.MacroPauseMax
	}

	return &BroadcastWorker{
		config:       config,
		manager:      manager,
		inputChan:    make(chan BroadcastJob, config.MaxWorkersGlobal*2),
		empresaCount: make(map[string]int),
	}
}

// SetRepo inyecta el repositorio de persistencia. Debe llamarse antes de Start.
func (w *BroadcastWorker) SetRepo(repo storage.JobQueueRepository) {
	w.repo = repo
}

func (w *BroadcastWorker) Start(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		w.wg.Add(1)
		go w.workerLoop()
	}
}

func (w *BroadcastWorker) workerLoop() {
	defer w.wg.Done()
	for job := range w.inputChan {
		w.processJob(context.Background(), job)
		w.activeWorkersMu.Lock()
		w.activeWorkers--
		w.activeWorkersMu.Unlock()
	}
}

func (w *BroadcastWorker) processJob(ctx context.Context, job BroadcastJob) {
	ruc := NormalizeAccountID(job.RUCEmpresa)
	w.acquireEmpresaSlot(ruc)
	defer w.releaseEmpresaSlot(ruc)

	// Marcar job como running en DB
	if w.repo != nil && job.JobID > 0 {
		_ = w.repo.UpdateStatus(ctx, job.JobID, domain.JobStatusRunning, nil)

		// Heartbeat goroutine: actualiza last_heartbeat cada 30s mientras corre
		hbCtx, hbCancel := context.WithCancel(ctx)
		defer hbCancel()
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					_ = w.repo.Heartbeat(hbCtx, job.JobID)
				case <-hbCtx.Done():
					return
				}
			}
		}()
	}

	// Si no hay manager o cliente conectado, marcar todo como fallido
	if w.manager == nil {
		w.failAllItems(ctx, job, ErrClientNotConnected.Error())
		return
	}
	client, ok := w.manager.Get(job.AccountID)
	if !ok || client == nil || !client.IsConnected() {
		w.failAllItems(ctx, job, ErrClientNotConnected.Error())
		return
	}

	// Preparar adjuntos una sola vez
	prepCtx, prepCancel := context.WithTimeout(ctx, 30*time.Second)
	prepared, err := prepareAttachments(prepCtx, client, job.Attachments)
	prepCancel()
	if err != nil {
		w.failAllItems(ctx, job, err.Error())
		return
	}

	// Procesar con batching + delays anti-ban
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	batches := SplitIntoBatches(job.Items, w.config.BatchSizeMin, w.config.BatchSizeMax, rng)

	// Construir mapa de índice → ItemID para persistencia
	itemIDByIndex := make(map[int]int64, len(job.ItemIDs))
	for i, id := range job.ItemIDs {
		itemIDByIndex[i] = id
	}

	msgCount := 0
	jobFailed := false

	for batchIdx, batch := range batches {
		if ctx.Err() != nil {
			break
		}

		for posInBatch, item := range batch {
			if ctx.Err() != nil {
				break
			}

			// Índice global del item en job.Items
			globalIdx := msgCount

			err := w.processItemWithRetry(item, client, job.AccountID, prepared)

			itemStatus := domain.JobItemSent
			errText := ""
			if err != nil {
				itemStatus = domain.JobItemFailed
				errText = err.Error()
				jobFailed = true
			}

			// Persistir resultado del item
			if w.repo != nil {
				if itemID, found := itemIDByIndex[globalIdx]; found {
					_ = w.repo.UpdateItemStatus(ctx, itemID, itemStatus, errText)
				}
			}

			msgCount++

			// Macro-pausa cada MacroPauseEvery mensajes (excepto al final)
			if w.config.MacroPauseEvery > 0 && msgCount%w.config.MacroPauseEvery == 0 && msgCount < len(job.Items) {
				d := RandDuration(rng, w.config.MacroPauseMin, w.config.MacroPauseMax)
				if err := SleepWithContext(ctx, d); err != nil {
					break
				}
			}

			// Delay intra-batch (no después del último item del batch)
			if posInBatch < len(batch)-1 {
				d := RandDuration(rng, w.config.IntraBatchDelayMin, w.config.IntraBatchDelayMax)
				if err := SleepWithContext(ctx, d); err != nil {
					break
				}
			}
		}

		// Delay inter-batch (no después del último batch)
		if batchIdx < len(batches)-1 {
			d := RandDuration(rng, w.config.InterBatchDelayMin, w.config.InterBatchDelayMax)
			if err := SleepWithContext(ctx, d); err != nil {
				break
			}
		}
	}

	// Actualizar estado final del job
	if w.repo != nil && job.JobID > 0 {
		finalStatus := domain.JobStatusCompleted
		if jobFailed {
			finalStatus = domain.JobStatusFailed
		}
		now := time.Now()
		_ = w.repo.UpdateStatus(ctx, job.JobID, finalStatus, &now)
	}
}

func (w *BroadcastWorker) failAllItems(ctx context.Context, job BroadcastJob, errMsg string) {
	if w.repo != nil {
		for i, itemID := range job.ItemIDs {
			_ = w.repo.UpdateItemStatus(ctx, itemID, domain.JobItemFailed, errMsg)
			_ = i
		}
		if job.JobID > 0 {
			now := time.Now()
			_ = w.repo.UpdateStatus(ctx, job.JobID, domain.JobStatusFailed, &now)
		}
	}
}

func (w *BroadcastWorker) processItemWithRetry(item domain.BroadcastItem, client *wa.Client, accountID string, prepared []preparedAttachment) error {
	var lastErr error
	for attempt := 0; attempt <= w.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(w.config.RetryDelay)
		}
		err := w.processItem(item, client, accountID, prepared)
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

func (w *BroadcastWorker) processItem(item domain.BroadcastItem, client *wa.Client, accountID string, prepared []preparedAttachment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return sendPreparedMessage(ctx, NewModuleLogger("WA-BROADCAST"), nil, client, accountID, item.Destino, item.Mensaje, prepared, "")
}

func (w *BroadcastWorker) acquireEmpresaSlot(ruc string) {
	w.empresaMu.Lock()
	defer w.empresaMu.Unlock()
	for w.empresaCount[ruc] >= w.config.MaxWorkersPerEmpresa {
		w.empresaMu.Unlock()
		time.Sleep(100 * time.Millisecond)
		w.empresaMu.Lock()
	}
	w.empresaCount[ruc]++
}

func (w *BroadcastWorker) releaseEmpresaSlot(ruc string) {
	w.empresaMu.Lock()
	defer w.empresaMu.Unlock()
	if w.empresaCount[ruc] > 0 {
		w.empresaCount[ruc]--
	}
}

func (w *BroadcastWorker) Submit(job BroadcastJob) error {
	w.activeWorkersMu.Lock()
	if w.stopped {
		w.activeWorkersMu.Unlock()
		return ErrPoolClosed
	}
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
		_ = w.Submit(job)
	}()
}

func (w *BroadcastWorker) Shutdown() {
	w.activeWorkersMu.Lock()
	if w.stopped {
		w.activeWorkersMu.Unlock()
		return
	}
	w.stopped = true
	close(w.inputChan)
	w.activeWorkersMu.Unlock()
	w.wg.Wait()
}

// Config retorna la configuración actual del worker.
func (w *BroadcastWorker) Config() WorkerConfig {
	return w.config
}

// ── helpers exportados para tests ────────────────────────────────────────────

// SplitIntoBatches divide items en grupos de tamaño [minSize, maxSize] usando rng local.
func SplitIntoBatches(items []domain.BroadcastItem, minSize, maxSize int, rng *rand.Rand) [][]domain.BroadcastItem {
	if minSize <= 0 || maxSize < minSize {
		minSize, maxSize = 3, 4
	}
	var batches [][]domain.BroadcastItem
	remaining := items
	for len(remaining) > 0 {
		size := minSize + rng.Intn(maxSize-minSize+1)
		if size > len(remaining) {
			size = len(remaining)
		}
		batches = append(batches, remaining[:size])
		remaining = remaining[size:]
	}
	return batches
}

// RandDuration retorna duración aleatoria en [min, max] usando rng local.
func RandDuration(rng *rand.Rand, min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}
	return min + time.Duration(rng.Int63n(int64(max-min)))
}

// SleepWithContext duerme d o retorna ctx.Err() si el contexto se cancela.
func SleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	for _, marker := range []string{"timeout", "connection", "reset", "refused", "temporary"} {
		if strings.Contains(s, marker) {
			return true
		}
	}
	return false
}

var ErrPoolFull = &BroadcastError{Message: "worker pool is full"}
var ErrPoolClosed = &BroadcastError{Message: "worker pool is closed"}

type BroadcastError struct {
	Message string
}

func (e *BroadcastError) Error() string { return e.Message }
