package whatsapp

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"wsapi/internal/domain"
)

func TestSplitIntoBatches_NoItemLost(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	items := make([]domain.BroadcastItem, 15)
	for i := range items {
		items[i] = domain.BroadcastItem{Destino: "51987000000", Mensaje: "test"}
	}

	batches := SplitIntoBatches(items, 3, 4, rng)

	totalCount := 0
	for _, batch := range batches {
		totalCount += len(batch)
	}

	if totalCount != len(items) {
		t.Errorf("expected %d items total, got %d", len(items), totalCount)
	}
}

func TestSplitIntoBatches_SizeRange(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	items := make([]domain.BroadcastItem, 50)
	for i := range items {
		items[i] = domain.BroadcastItem{Destino: "51987000000", Mensaje: "test"}
	}

	batches := SplitIntoBatches(items, 3, 4, rng)

	for i, batch := range batches {
		if len(batch) < 3 || len(batch) > 4 {
			// El último lote puede ser menor si no quedan suficientes items
			if i == len(batches)-1 && len(batch) <= 4 {
				continue
			}
			t.Errorf("batch %d has invalid size: %d", i, len(batch))
		}
	}
}

func TestRandDuration_Bounds(t *testing.T) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	min := 2 * time.Second
	max := 5 * time.Second

	for i := 0; i < 1000; i++ {
		d := RandDuration(rng, min, max)
		if d < min || d >= max {
			t.Errorf("duration %v out of bounds [%v, %v]", d, min, max)
		}
	}
}

func TestSleepWithContext_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := SleepWithContext(ctx, 1*time.Second)
	if err == nil {
		t.Error("expected context.Canceled error, got nil")
	}
}

func TestEstimateBroadcastSeconds(t *testing.T) {
	cfg := domain.BroadcastTimingConfig{
		BatchSizeMin:       3,
		BatchSizeMax:       4,
		IntraBatchDelayMin: 1 * time.Second,
		IntraBatchDelayMax: 1 * time.Second,
		InterBatchDelayMin: 2 * time.Second,
		InterBatchDelayMax: 2 * time.Second,
		MacroPauseEvery:    10,
		MacroPauseMin:      5 * time.Second,
		MacroPauseMax:      5 * time.Second,
	}

	// 10 destinatarios.
	// Tamaño del lote medio: (3+4)/2 = 3.5.
	// Lotes calculados: Ceil(10 / 3.5) = Ceil(2.85) = 3 lotes.
	// Delays intra-batch: 10 - 3 = 7 delays de 1s = 7s.
	// Delays inter-batch: 3 - 1 = 2 delays de 2s = 4s.
	// Macro-pausa: Floor(10 / 10) = 1 pausa de 5s = 5s.
	// Total: 7s + 4s + 5s = 16s.
	est := domain.EstimateBroadcastSeconds(10, cfg)
	if est != 16 {
		t.Errorf("expected 16s, got %ds", est)
	}
}
