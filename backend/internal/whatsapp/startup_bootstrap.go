package whatsapp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"wsapi/internal/domain"
	"wsapi/internal/metrics"
	"wsapi/internal/storage"
)

type StartupBootstrapConfig struct {
	MaxConcurrency int
	MaxRetries     int
	RetryDelay     time.Duration
}

var DefaultStartupBootstrapConfig = StartupBootstrapConfig{
	MaxConcurrency: 4,
	MaxRetries:     2,
	RetryDelay:     1200 * time.Millisecond,
}

type StartupBootstrapSummary struct {
	TotalTelefonos       int
	ActivosEnDB          int
	RuntimeActivos       int
	MismatchesDetectados int
	IntentosStart        int
	ErroresStart         int
	Duracion             time.Duration
}

type StartupBootstrapper struct {
	manager       *Manager
	sessionStore  *storage.SessionStore
	telefonoStore *storage.TelefonoStore
	config        StartupBootstrapConfig
}

func NewStartupBootstrapper(manager *Manager, sessionStore *storage.SessionStore, telefonoStore *storage.TelefonoStore, config StartupBootstrapConfig) *StartupBootstrapper {
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = DefaultStartupBootstrapConfig.MaxConcurrency
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = DefaultStartupBootstrapConfig.MaxRetries
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = DefaultStartupBootstrapConfig.RetryDelay
	}
	return &StartupBootstrapper{
		manager:       manager,
		sessionStore:  sessionStore,
		telefonoStore: telefonoStore,
		config:        config,
	}
}

func (b *StartupBootstrapper) Run(ctx context.Context) StartupBootstrapSummary {
	startedAt := time.Now()
	summary := StartupBootstrapSummary{}

	if b == nil || b.manager == nil || b.telefonoStore == nil {
		return summary
	}

	telefonos, err := b.telefonoStore.ListAll()
	if err != nil {
		fmt.Printf("[WARN] startup bootstrap: no se pudieron listar telefonos: %v\n", err)
		return summary
	}

	summary.TotalTelefonos = len(telefonos)

	type candidate struct {
		id        int64
		accountID string
	}

	candidates := make([]candidate, 0, len(telefonos))

	for _, t := range telefonos {
		accountID := NormalizeAccountID(t.NumeroCompleto)
		if accountID == "" {
			continue
		}

		client, ok := b.manager.Get(accountID)
		runtimeActive := ok && client != nil && client.IsConnected()
		if runtimeActive {
			summary.RuntimeActivos++
		}

		if t.Status == domain.TelefonoStatusActive {
			summary.ActivosEnDB++
			if runtimeActive {
				if b.sessionStore != nil {
					b.sessionStore.SetActive(accountID)
				}
				continue
			}

			summary.MismatchesDetectados++
			candidates = append(candidates, candidate{id: t.ID, accountID: accountID})
			if b.sessionStore != nil {
				b.sessionStore.SetInitializing(accountID)
				b.sessionStore.AppendEvent(accountID, "initializing", "bootstrap_restart")
			}
			continue
		}

		if runtimeActive {
			summary.MismatchesDetectados++
			if err := b.telefonoStore.SetConnected(t.ID); err != nil {
				fmt.Printf("[WARN] startup bootstrap: no se pudo reconciliar telefono %d a active: %v\n", t.ID, err)
			}
			if b.sessionStore != nil {
				b.sessionStore.SetActive(accountID)
			}
		}
	}

	if len(candidates) > 0 {
		sem := make(chan struct{}, b.config.MaxConcurrency)
		var wg sync.WaitGroup
		var mu sync.Mutex

		for _, c := range candidates {
			if ctx.Err() != nil {
				break
			}

			wg.Add(1)
			sem <- struct{}{}

			go func(c candidate) {
				defer wg.Done()
				defer func() { <-sem }()

				events, unsubscribe, attempts, err := b.startSessionWithRetry(ctx, c.accountID)
				mu.Lock()
				summary.IntentosStart += attempts
				if err != nil {
					summary.ErroresStart++
				}
				mu.Unlock()

				if err != nil {
					fmt.Printf("[WARN] startup bootstrap: no se pudo iniciar sesion %s: %v\n", c.accountID, err)
					if b.sessionStore != nil {
						b.sessionStore.SetDisconnected(c.accountID, "startup_start_failed")
						b.sessionStore.AppendEvent(c.accountID, "disconnected", "startup_start_failed")
					}
					if setErr := b.telefonoStore.SetDisconnected(c.id); setErr != nil {
						fmt.Printf("[WARN] startup bootstrap: no se pudo marcar disconnected telefono %d: %v\n", c.id, setErr)
					}
					return
				}

				go func(ch <-chan SessionEvent, unsub func()) {
					defer unsub()
					for {
						select {
						case <-ctx.Done():
							return
						case _, ok := <-ch:
							if !ok {
								return
							}
						}
					}
				}(events, unsubscribe)
			}(c)
		}

		wg.Wait()
	}

	summary.Duracion = time.Since(startedAt)
	metrics.IncrementStartupBootstrapRuns()
	metrics.AddStartupBootstrapMismatches(summary.MismatchesDetectados)
	metrics.AddStartupBootstrapStartAttempts(summary.IntentosStart)
	metrics.AddStartupBootstrapStartErrors(summary.ErroresStart)
	metrics.SetStartupBootstrapLastDurationMs(summary.Duracion.Milliseconds())
	return summary
}

func (b *StartupBootstrapper) startSessionWithRetry(ctx context.Context, accountID string) (<-chan SessionEvent, func(), int, error) {
	var lastErr error
	attempts := 0
	for attempt := 0; attempt <= b.config.MaxRetries; attempt++ {
		attempts++
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, nil, attempts, ctx.Err()
			case <-time.After(b.config.RetryDelay):
			}
		}

		events, unsubscribe, err := StartSession(b.manager, accountID)
		if err == nil {
			return events, unsubscribe, attempts, nil
		}
		lastErr = err
	}
	return nil, nil, attempts, lastErr
}
