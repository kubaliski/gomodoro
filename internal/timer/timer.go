package timer

import (
	"time"
)

// Timer representa el temporizador
type Timer struct {
	Duration  time.Duration
	Remaining time.Duration
	IsRunning bool
	IsPaused  bool
}

// NewTimer crea un nuevo Timer con una duracción especificada
func NewTimer(duration time.Duration) *Timer {
	return &Timer{
		Duration:  duration,
		Remaining: duration,
		IsRunning: false,
		IsPaused:  false,
	}
}

// Start inicia el timer
func (t *Timer) Start() {
	t.IsRunning = true
	t.IsPaused = false
}

// Pause pausa el timer
func (t *Timer) Pause() {
	t.IsPaused = true
}

// Resume reanuda el timer
func (t *Timer) Resume() {
	t.IsPaused = false
}

// Stop detiene el timer
func (t *Timer) Stop() {
	t.IsRunning = false
	t.IsPaused = false
}

// IsFinished verifica si el timer ha terminado
func (t *Timer) IsFinished() bool {
	return t.Remaining <= 0
}
