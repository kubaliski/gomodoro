package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kubaliski/pomodoro-core/config"
	"github.com/kubaliski/pomodoro-core/events"
	"github.com/kubaliski/pomodoro-core/stats"
	"github.com/kubaliski/pomodoro-core/timer"
)

// State representa el estado actual del engine
type State string

const (
	StateIdle    State = "idle"
	StateRunning State = "running"
	StatePaused  State = "paused"
	StateStopped State = "stopped"
)

// SessionType representa el tipo de sesión actual
type SessionType string

const (
	SessionWork       SessionType = "work"
	SessionShortBreak SessionType = "short_break"
	SessionLongBreak  SessionType = "long_break"
)

// Engine es el motor principal del pomodoro, thread-safe e independiente de UI
type Engine struct {
	mu sync.RWMutex

	// Configuración inmutable
	config *config.Config

	// Estado mutable
	state          State
	currentSession SessionType
	pomodoroCount  int
	isRunning      bool

	// Componentes
	currentTimer *timer.Timer
	statsManager *stats.SessionStats
	eventBus     *events.EventBus

	// Control de tiempo
	sessionStartTime time.Time

	// Control de contexto
	ctx    context.Context
	cancel context.CancelFunc

	// Canales para coordinación
	commandChan chan command
	tickerDone  chan struct{}
}

// command representa comandos internos del engine
type command struct {
	action string
	data   interface{}
	result chan error
}

// EngineInterface define la interfaz pública del engine
type EngineInterface interface {
	Start(ctx context.Context) error
	StartFirstSession() error
	Stop() error
	Pause() error
	Resume() error
	Skip() error
	GetState() State
	GetCurrentSession() SessionType
	GetPomodoroCount() int
	IsRunning() bool
	GetStats() *stats.SessionStats
	GetEventBus() *events.EventBus
	GetConfig() *config.Config
}

// NewEngine crea una nueva instancia del motor de pomodoro
func NewEngine(cfg *config.Config) *Engine {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Validar configuración
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid configuration: %v", err))
	}

	return &Engine{
		config:         cfg.Clone(), // Usar copia para inmutabilidad
		state:          StateIdle,
		currentSession: SessionWork,
		pomodoroCount:  0,
		isRunning:      false,
		statsManager:   stats.NewSessionStats(),
		eventBus:       events.NewEventBus(),
		commandChan:    make(chan command, 10),
		tickerDone:     make(chan struct{}),
	}
}

// Implementación de EngineInterface

// Start inicia el motor en el contexto proporcionado
func (e *Engine) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isRunning {
		return nil // Ya está corriendo
	}

	// Configurar contexto
	e.ctx, e.cancel = context.WithCancel(ctx)
	e.isRunning = true
	e.state = StateIdle
	e.pomodoroCount = 0
	e.currentSession = SessionWork // Iniciar en trabajo

	// Iniciar goroutine principal pero SIN empezar sesión automáticamente
	go e.runEventLoop()

	// Emitir evento de inicio
	e.eventBus.Publish(events.EngineStarted, events.SessionEventData{
		SessionID:  fmt.Sprintf("session_%d", time.Now().Unix()),
		StartTime:  time.Now(),
		ConfigUsed: e.config,
	})

	return nil
}

// StartFirstSession inicia manualmente la primera sesión
func (e *Engine) StartFirstSession() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isRunning {
		return fmt.Errorf("engine not running")
	}

	if e.currentTimer != nil {
		return nil // Ya hay una sesión corriendo
	}

	// Iniciar primera sesión (trabajo)
	go e.startNextSession()
	return nil
}

// Stop detiene el motor completamente
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isRunning {
		return nil
	}

	e.isRunning = false
	e.state = StateStopped

	if e.currentTimer != nil {
		e.currentTimer.Stop()
	}

	if e.cancel != nil {
		e.cancel()
	}

	// Emitir evento de parada
	e.eventBus.Publish(events.EngineStopped, events.SessionEventData{
		EndTime:   time.Now(),
		TotalTime: e.statsManager.GetSessionDuration(),
	})

	return nil
}

// Pause pausa el timer actual
func (e *Engine) Pause() error {
	return e.sendCommand("pause", nil)
}

// Resume reanuda el timer pausado
func (e *Engine) Resume() error {
	return e.sendCommand("resume", nil)
}

// Skip salta la sesión actual
func (e *Engine) Skip() error {
	return e.sendCommand("skip", nil)
}

// GetState retorna el estado actual del engine
func (e *Engine) GetState() State {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state
}

// GetCurrentSession retorna el tipo de sesión actual
func (e *Engine) GetCurrentSession() SessionType {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.currentSession
}

// GetPomodoroCount retorna el número de pomodoros completados
func (e *Engine) GetPomodoroCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.pomodoroCount
}

// IsRunning indica si el engine está ejecutándose
func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isRunning
}

// GetStats retorna el gestor de estadísticas
func (e *Engine) GetStats() *stats.SessionStats {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.statsManager
}

// GetEventBus retorna el bus de eventos
func (e *Engine) GetEventBus() *events.EventBus {
	return e.eventBus
}

// GetConfig retorna una copia de la configuración
func (e *Engine) GetConfig() *config.Config {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config.Clone()
}

// Métodos privados

// runEventLoop es el bucle principal del engine
func (e *Engine) runEventLoop() {
	defer func() {
		if r := recover(); r != nil {
			e.eventBus.Publish(events.ErrorOccurred, events.ErrorEventData{
				Message: fmt.Sprintf("Engine panic: %v", r),
				Code:    "ENGINE_PANIC",
				Source:  "engine.runEventLoop",
			})
		}
	}()

	// Ticker para actualizaciones del timer
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return

		case cmd := <-e.commandChan:
			e.handleCommand(cmd)

		case <-ticker.C:
			e.handleTick()

		case <-e.tickerDone:
			return
		}
	}
}

// sendCommand envía un comando al engine y espera respuesta
func (e *Engine) sendCommand(action string, data interface{}) error {
	if !e.IsRunning() {
		return fmt.Errorf("engine is not running")
	}

	cmd := command{
		action: action,
		data:   data,
		result: make(chan error, 1),
	}

	select {
	case e.commandChan <- cmd:
		return <-cmd.result
	case <-e.ctx.Done():
		return e.ctx.Err()
	}
}

// handleCommand procesa comandos del engine
func (e *Engine) handleCommand(cmd command) {
	var err error

	switch cmd.action {
	case "pause":
		err = e.pauseCurrentTimer()
	case "resume":
		err = e.resumeCurrentTimer()
	case "skip":
		err = e.skipCurrentTimer()
	default:
		err = fmt.Errorf("unknown command: %s", cmd.action)
	}

	select {
	case cmd.result <- err:
	default:
	}
}

// handleTick maneja las actualizaciones del timer
func (e *Engine) handleTick() {
	e.mu.RLock()
	currentTimer := e.currentTimer
	e.mu.RUnlock()

	if currentTimer == nil {
		return
	}

	// Actualizar timer
	snapshot := currentTimer.Tick()

	// Emitir evento de tick
	e.eventBus.Publish(events.TimerTick, e.createTimerEventData(snapshot))

	// Verificar estados
	if currentTimer.IsFinished() {
		e.handleTimerCompleted()
	} else if currentTimer.IsSkipped() {
		e.handleTimerSkipped()
	}
}

// startNextSession inicia la siguiente sesión
func (e *Engine) startNextSession() {
	e.mu.Lock()

	if !e.isRunning {
		e.mu.Unlock()
		return
	}

	// Determinar tipo y duración de sesión
	var duration time.Duration
	var nextSessionType SessionType

	// Si es la primera sesión (no hay timer previo), empezar con trabajo
	if e.currentTimer == nil {
		nextSessionType = SessionWork
		duration = e.config.WorkDuration
	} else if e.currentSession == SessionWork {
		// Trabajo completado, siguiente es descanso
		e.pomodoroCount++
		var isLong bool
		duration, isLong = e.config.GetNextBreakType(e.pomodoroCount)
		if isLong {
			nextSessionType = SessionLongBreak
		} else {
			nextSessionType = SessionShortBreak
		}
	} else {
		// Descanso completado, siguiente es trabajo
		nextSessionType = SessionWork
		duration = e.config.WorkDuration
	}

	e.currentSession = nextSessionType
	e.updateStateFromSession()
	e.sessionStartTime = time.Now()

	// Crear nuevo timer
	e.currentTimer = timer.NewTimer(duration)
	e.currentTimer.Start()

	e.mu.Unlock()

	// Emitir eventos apropiados
	e.emitSessionStartedEvent(nextSessionType, duration)
	e.eventBus.Publish(events.TimerStarted, e.createTimerEventData(e.currentTimer.GetSnapshot()))
}

// emitSessionStartedEvent emite el evento apropiado según el tipo de sesión
func (e *Engine) emitSessionStartedEvent(sessionType SessionType, duration time.Duration) {
	switch sessionType {
	case SessionWork:
		e.eventBus.Publish(events.PomodoroStarted, events.PomodoroEventData{
			Number:    e.pomodoroCount + 1, // +1 porque aún no se ha completado
			Duration:  duration,
			StartTime: e.sessionStartTime,
		})
	case SessionShortBreak, SessionLongBreak:
		e.eventBus.Publish(events.BreakStarted, events.BreakEventData{
			Type:        e.getBreakTypeString(sessionType),
			Duration:    duration,
			StartTime:   e.sessionStartTime,
			IsLongBreak: sessionType == SessionLongBreak,
		})
	}
}

// handleTimerCompleted maneja cuando un timer se completa
func (e *Engine) handleTimerCompleted() {
	e.mu.Lock()
	sessionEndTime := time.Now()
	actualTime := sessionEndTime.Sub(e.sessionStartTime)
	currentSession := e.currentSession
	duration := e.currentTimer.GetDuration()
	e.mu.Unlock()

	// Actualizar estadísticas
	switch currentSession {
	case SessionWork:
		e.statsManager.AddCompletedPomodoro(duration, actualTime, e.sessionStartTime, sessionEndTime)
	case SessionShortBreak, SessionLongBreak:
		breakType := e.getBreakTypeString(currentSession)
		e.statsManager.AddCompletedBreak(breakType, duration, actualTime, e.sessionStartTime, sessionEndTime)
	}

	// Emitir eventos
	e.eventBus.Publish(events.TimerCompleted, e.createTimerEventData(e.currentTimer.GetSnapshot()))
	e.eventBus.Publish(events.StatsUpdated, e.createStatsEventData())

	// Emitir evento específico de sesión
	e.emitSessionCompletedEvent(currentSession, duration, actualTime, sessionEndTime)

	// Continuar con siguiente sesión
	go e.startNextSession()
}

// handleTimerSkipped maneja cuando un timer es saltado
func (e *Engine) handleTimerSkipped() {
	e.mu.Lock()
	sessionEndTime := time.Now()
	actualTime := sessionEndTime.Sub(e.sessionStartTime)
	currentSession := e.currentSession
	duration := e.currentTimer.GetDuration()
	e.mu.Unlock()

	// Actualizar estadísticas
	switch currentSession {
	case SessionWork:
		e.statsManager.AddSkippedPomodoro(duration, actualTime, e.sessionStartTime, sessionEndTime)
	case SessionShortBreak, SessionLongBreak:
		breakType := e.getBreakTypeString(currentSession)
		e.statsManager.AddSkippedBreak(breakType, duration, actualTime, e.sessionStartTime, sessionEndTime)
	}

	// Emitir eventos
	e.eventBus.Publish(events.TimerSkipped, e.createTimerEventData(e.currentTimer.GetSnapshot()))
	e.eventBus.Publish(events.StatsUpdated, e.createStatsEventData())

	// Emitir evento específico de sesión
	e.emitSessionSkippedEvent(currentSession, duration, actualTime, sessionEndTime)

	// Continuar con siguiente sesión
	go e.startNextSession()
}

// emitSessionCompletedEvent emite evento de sesión completada
func (e *Engine) emitSessionCompletedEvent(sessionType SessionType, duration, actualTime time.Duration, endTime time.Time) {
	switch sessionType {
	case SessionWork:
		e.eventBus.Publish(events.PomodoroCompleted, events.PomodoroEventData{
			Number:     e.pomodoroCount,
			Duration:   duration,
			ActualTime: actualTime,
			StartTime:  e.sessionStartTime,
			EndTime:    endTime,
		})
	case SessionShortBreak, SessionLongBreak:
		e.eventBus.Publish(events.BreakCompleted, events.BreakEventData{
			Type:        e.getBreakTypeString(sessionType),
			Duration:    duration,
			ActualTime:  actualTime,
			StartTime:   e.sessionStartTime,
			EndTime:     endTime,
			IsLongBreak: sessionType == SessionLongBreak,
		})
	}
}

// emitSessionSkippedEvent emite evento de sesión saltada
func (e *Engine) emitSessionSkippedEvent(sessionType SessionType, duration, actualTime time.Duration, endTime time.Time) {
	switch sessionType {
	case SessionWork:
		e.eventBus.Publish(events.PomodoroSkipped, events.PomodoroEventData{
			Number:     e.pomodoroCount,
			Duration:   duration,
			ActualTime: actualTime,
			StartTime:  e.sessionStartTime,
			EndTime:    endTime,
		})
	case SessionShortBreak, SessionLongBreak:
		e.eventBus.Publish(events.BreakSkipped, events.BreakEventData{
			Type:        e.getBreakTypeString(sessionType),
			Duration:    duration,
			ActualTime:  actualTime,
			StartTime:   e.sessionStartTime,
			EndTime:     endTime,
			IsLongBreak: sessionType == SessionLongBreak,
		})
	}
}

// pauseCurrentTimer pausa el timer actual
func (e *Engine) pauseCurrentTimer() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.currentTimer != nil && e.currentTimer.IsRunning() && !e.currentTimer.IsPaused() {
		e.currentTimer.Pause()
		e.state = StatePaused
		e.eventBus.Publish(events.TimerPaused, e.createTimerEventData(e.currentTimer.GetSnapshot()))
	}
	return nil
}

// resumeCurrentTimer reanuda el timer pausado
func (e *Engine) resumeCurrentTimer() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.currentTimer != nil && e.currentTimer.IsPaused() {
		e.currentTimer.Resume()
		e.updateStateFromSession()
		e.eventBus.Publish(events.TimerResumed, e.createTimerEventData(e.currentTimer.GetSnapshot()))
	}
	return nil
}

// skipCurrentTimer salta el timer actual
func (e *Engine) skipCurrentTimer() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.currentTimer != nil && (e.currentTimer.IsRunning() || e.currentTimer.IsPaused()) {
		e.currentTimer.Skip()
	}
	return nil
}

// updateStateFromSession actualiza el estado basado en la sesión actual
func (e *Engine) updateStateFromSession() {
	switch e.currentSession {
	case SessionWork:
		e.state = StateRunning
	case SessionShortBreak, SessionLongBreak:
		e.state = StateRunning
	}
}

// getBreakTypeString convierte SessionType a string
func (e *Engine) getBreakTypeString(sessionType SessionType) string {
	switch sessionType {
	case SessionLongBreak:
		return "DESCANSO LARGO"
	default:
		return "DESCANSO"
	}
}

// createTimerEventData crea datos de evento del timer
func (e *Engine) createTimerEventData(snapshot timer.TimerSnapshot) events.TimerEventData {
	var stateStr string
	switch e.currentSession {
	case SessionWork:
		stateStr = "TRABAJO"
	case SessionShortBreak:
		stateStr = "DESCANSO"
	case SessionLongBreak:
		stateStr = "DESCANSO LARGO"
	}

	var statusStr string
	switch snapshot.State {
	case timer.StatePaused:
		statusStr = "PAUSED"
	case timer.StateRunning:
		statusStr = "RUNNING"
	default:
		statusStr = "STOPPED"
	}

	return events.TimerEventData{
		Remaining:    snapshot.Remaining,
		Total:        snapshot.Duration,
		State:        stateStr,
		Status:       statusStr,
		Progress:     snapshot.Progress,
		SessionCount: e.pomodoroCount,
	}
}

// createStatsEventData crea datos de evento de estadísticas
func (e *Engine) createStatsEventData() events.StatsEventData {
	snapshot := e.statsManager.GetSnapshot()
	return events.StatsEventData{
		PomodorosCompleted: snapshot.PomodorosCompleted,
		PomodorosSkipped:   snapshot.PomodorosSkipped,
		BreaksCompleted:    snapshot.BreaksCompleted,
		BreaksSkipped:      snapshot.BreaksSkipped,
		CurrentStreak:      snapshot.CurrentStreak,
		BestStreak:         snapshot.BestStreak,
		TotalWorkTime:      snapshot.TotalWorkTime,
		TotalBreakTime:     snapshot.TotalBreakTime,
		SessionDuration:    snapshot.SessionDuration,
		WorkEfficiency:     snapshot.WorkEfficiency,
	}
}
