package timer

import (
	"context"
	"sync"
	"time"
)

// State representa el estado del timer
type State string

const (
	StateIdle    State = "idle"
	StateRunning State = "running"
	StatePaused  State = "paused"
	StateSkipped State = "skipped"
	StateDone    State = "done"
)

// Timer representa un temporizador thread-safe
type Timer struct {
	mu sync.RWMutex

	// Configuración inmutable
	duration time.Duration

	// Estado mutable
	remaining   time.Duration
	state       State
	startedAt   time.Time
	pausedAt    time.Time
	totalPaused time.Duration

	// Control de contexto
	ctx    context.Context
	cancel context.CancelFunc

	// Canales para comunicación
	tickChan chan time.Duration
	doneChan chan struct{}
	skipChan chan struct{}
}

// TimerSnapshot representa una instantánea inmutable del estado del timer
type TimerSnapshot struct {
	Duration      time.Duration
	Remaining     time.Duration
	State         State
	Progress      float64
	StartedAt     time.Time
	ElapsedActive time.Duration
	TotalPaused   time.Duration
}

// NewTimer crea un nuevo timer con la duración especificada
func NewTimer(duration time.Duration) *Timer {
	ctx, cancel := context.WithCancel(context.Background())

	return &Timer{
		duration:  duration,
		remaining: duration,
		state:     StateIdle,
		ctx:       ctx,
		cancel:    cancel,
		tickChan:  make(chan time.Duration, 1),
		doneChan:  make(chan struct{}, 1),
		skipChan:  make(chan struct{}, 1),
	}
}

// Start inicia el timer
func (t *Timer) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != StateIdle && t.state != StatePaused {
		return nil // Ya está corriendo o terminado
	}

	if t.state == StateIdle {
		t.startedAt = time.Now()
		t.totalPaused = 0
	} else if t.state == StatePaused {
		// Reanudar desde pausa
		pauseDuration := time.Since(t.pausedAt)
		t.totalPaused += pauseDuration
	}

	t.state = StateRunning
	return nil
}

// Pause pausa el timer
func (t *Timer) Pause() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state != StateRunning {
		return nil // No está corriendo
	}

	t.state = StatePaused
	t.pausedAt = time.Now()
	return nil
}

// Resume reanuda el timer pausado
func (t *Timer) Resume() error {
	return t.Start() // Start maneja la reanudación
}

// Skip salta el timer actual
func (t *Timer) Skip() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateRunning || t.state == StatePaused {
		t.state = StateSkipped
		select {
		case t.skipChan <- struct{}{}:
		default:
		}
	}
}

// Stop detiene el timer completamente
func (t *Timer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cancel()
	t.state = StateIdle
	t.remaining = t.duration
}

// Reset reinicia el timer a su estado inicial
func (t *Timer) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Cancelar contexto anterior y crear nuevo
	t.cancel()
	t.ctx, t.cancel = context.WithCancel(context.Background())

	t.remaining = t.duration
	t.state = StateIdle
	t.startedAt = time.Time{}
	t.pausedAt = time.Time{}
	t.totalPaused = 0
}

// Tick actualiza el timer (llamado cada segundo)
func (t *Timer) Tick() TimerSnapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.state == StateRunning && t.remaining > 0 {
		t.remaining -= time.Second

		// Notificar tick
		select {
		case t.tickChan <- t.remaining:
		default:
		}

		// Verificar si terminó
		if t.remaining <= 0 {
			t.state = StateDone
			select {
			case t.doneChan <- struct{}{}:
			default:
			}
		}
	}

	return t.createSnapshot()
}

// GetSnapshot retorna una instantánea actual del timer
func (t *Timer) GetSnapshot() TimerSnapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.createSnapshot()
}

// createSnapshot crea una instantánea (debe llamarse con lock)
func (t *Timer) createSnapshot() TimerSnapshot {
	var progress float64
	if t.duration > 0 {
		progress = float64(t.duration-t.remaining) / float64(t.duration)
	}

	var elapsedActive time.Duration
	if !t.startedAt.IsZero() {
		elapsed := time.Since(t.startedAt)
		elapsedActive = elapsed - t.totalPaused

		// Si está pausado, no incluir el tiempo de pausa actual
		if t.state == StatePaused {
			elapsedActive -= time.Since(t.pausedAt)
		}
	}

	return TimerSnapshot{
		Duration:      t.duration,
		Remaining:     t.remaining,
		State:         t.state,
		Progress:      progress,
		StartedAt:     t.startedAt,
		ElapsedActive: elapsedActive,
		TotalPaused:   t.totalPaused,
	}
}

// GetState retorna el estado actual del timer
func (t *Timer) GetState() State {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

// IsRunning verifica si el timer está corriendo
func (t *Timer) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state == StateRunning
}

// IsPaused verifica si el timer está pausado
func (t *Timer) IsPaused() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state == StatePaused
}

// IsFinished verifica si el timer ha terminado
func (t *Timer) IsFinished() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state == StateDone
}

// IsSkipped verifica si el timer fue saltado
func (t *Timer) IsSkipped() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state == StateSkipped
}

// GetDuration retorna la duración total del timer
func (t *Timer) GetDuration() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.duration
}

// GetRemaining retorna el tiempo restante
func (t *Timer) GetRemaining() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.remaining
}

// TickChan retorna el canal de ticks
func (t *Timer) TickChan() <-chan time.Duration {
	return t.tickChan
}

// DoneChan retorna el canal de finalización
func (t *Timer) DoneChan() <-chan struct{} {
	return t.doneChan
}

// SkipChan retorna el canal de skip
func (t *Timer) SkipChan() <-chan struct{} {
	return t.skipChan
}

// Context retorna el contexto del timer
func (t *Timer) Context() context.Context {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.ctx
}
