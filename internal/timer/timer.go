package timer

import (
	"time"
)

// Timer representa nuestro temporizador pomodoro
type Timer struct {
	Duration  time.Duration
	Remaining time.Duration
	IsRunning bool
	IsPaused  bool
	IsSkipped bool
}

// NewTimer crea un nuevo timer con la duración especificada
func NewTimer(duration time.Duration) *Timer {
	return &Timer{
		Duration:  duration,
		Remaining: duration,
		IsRunning: false,
		IsPaused:  false,
		IsSkipped: false,
	}
}

// Start inicia el timer
func (t *Timer) Start() {
	t.IsRunning = true
	t.IsPaused = false
	t.IsSkipped = false
}

// Pause pausa el timer
func (t *Timer) Pause() {
	if t.IsRunning {
		t.IsPaused = true
	}
}

// Resume reanuda el timer
func (t *Timer) Resume() {
	if t.IsRunning {
		t.IsPaused = false
	}
}

// Skip salta el timer actual
func (t *Timer) Skip() {
	t.IsSkipped = true
	t.IsRunning = false
}

// Stop detiene el timer
func (t *Timer) Stop() {
	t.IsRunning = false
	t.IsPaused = false
}

// Reset reinicia el timer a su duración original
func (t *Timer) Reset() {
	t.Remaining = t.Duration
	t.IsRunning = false
	t.IsPaused = false
	t.IsSkipped = false
}

// IsFinished verifica si el timer ha terminado
func (t *Timer) IsFinished() bool {
	return t.Remaining <= 0
}

// Tick reduce el tiempo en un segundo si no está pausado
func (t *Timer) Tick() {
	if t.IsRunning && !t.IsPaused && t.Remaining > 0 {
		t.Remaining -= time.Second
	}
}

// GetStatus retorna el estado actual del timer
func (t *Timer) GetStatus() string {
	if !t.IsRunning {
		return "DETENIDO"
	}
	if t.IsPaused {
		return "PAUSADO"
	}
	return "CORRIENDO"
}
