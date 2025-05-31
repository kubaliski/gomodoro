package events

import (
	"sync"
	"time"
)

// EventType define los tipos de eventos que el sistema puede emitir
type EventType string

const (
	// Eventos del Engine
	EngineStarted EventType = "engine_started"
	EngineStopped EventType = "engine_stopped"

	// Eventos del Timer
	TimerStarted   EventType = "timer_started"
	TimerTick      EventType = "timer_tick"
	TimerPaused    EventType = "timer_paused"
	TimerResumed   EventType = "timer_resumed"
	TimerCompleted EventType = "timer_completed"
	TimerSkipped   EventType = "timer_skipped"

	// Eventos de Session
	SessionStarted EventType = "session_started"
	SessionEnded   EventType = "session_ended"

	// Eventos de Pomodoro
	PomodoroStarted   EventType = "pomodoro_started"
	PomodoroCompleted EventType = "pomodoro_completed"
	PomodoroSkipped   EventType = "pomodoro_skipped"

	// Eventos de Break
	BreakStarted   EventType = "break_started"
	BreakCompleted EventType = "break_completed"
	BreakSkipped   EventType = "break_skipped"

	// Eventos de Stats
	StatsUpdated EventType = "stats_updated"

	// Eventos de Error
	ErrorOccurred EventType = "error_occurred"
)

// Event representa un evento emitido por el sistema
type Event struct {
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// EventHandler define la interfaz para manejar eventos
type EventHandler interface {
	HandleEvent(event Event)
}

// EventHandlerFunc permite usar funciones como EventHandler
type EventHandlerFunc func(event Event)

func (f EventHandlerFunc) HandleEvent(event Event) {
	f(event)
}

// EventBus maneja la distribución de eventos de forma thread-safe
type EventBus struct {
	mu       sync.RWMutex
	handlers map[EventType][]EventHandler
	global   []EventHandler
}

// NewEventBus crea un nuevo bus de eventos
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]EventHandler),
		global:   make([]EventHandler, 0),
	}
}

// Subscribe registra un handler para un tipo específico de evento
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// SubscribeFunc registra una función como handler para un tipo específico de evento
func (eb *EventBus) SubscribeFunc(eventType EventType, handlerFunc func(event Event)) {
	eb.Subscribe(eventType, EventHandlerFunc(handlerFunc))
}

// SubscribeGlobal registra un handler que recibe todos los eventos
func (eb *EventBus) SubscribeGlobal(handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.global = append(eb.global, handler)
}

// SubscribeGlobalFunc registra una función como handler global
func (eb *EventBus) SubscribeGlobalFunc(handlerFunc func(event Event)) {
	eb.SubscribeGlobal(EventHandlerFunc(handlerFunc))
}

// Publish emite un evento a todos los handlers suscritos
func (eb *EventBus) Publish(eventType EventType, data interface{}) {
	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// Enviar a handlers globales
	for _, handler := range eb.global {
		go handler.HandleEvent(event)
	}

	// Enviar a handlers específicos del tipo
	if handlers, exists := eb.handlers[eventType]; exists {
		for _, handler := range handlers {
			go handler.HandleEvent(event)
		}
	}
}

// Unsubscribe remueve un handler específico
func (eb *EventBus) Unsubscribe(eventType EventType, targetHandler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if handlers, exists := eb.handlers[eventType]; exists {
		for i, handler := range handlers {
			// Comparación por dirección de memoria
			if &handler == &targetHandler {
				eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
}

// Clear limpia todos los handlers
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers = make(map[EventType][]EventHandler)
	eb.global = make([]EventHandler, 0)
}

// GetSubscriberCount retorna el número de suscriptores para un tipo de evento
func (eb *EventBus) GetSubscriberCount(eventType EventType) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.handlers[eventType])
}

// GetGlobalSubscriberCount retorna el número de suscriptores globales
func (eb *EventBus) GetGlobalSubscriberCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.global)
}

// Event Data Types - Estructuras específicas para cada tipo de evento

// TimerEventData contiene datos específicos de eventos del timer
type TimerEventData struct {
	Remaining    time.Duration `json:"remaining"`
	Total        time.Duration `json:"total"`
	State        string        `json:"state"`    // "TRABAJO", "DESCANSO", "DESCANSO LARGO"
	Status       string        `json:"status"`   // "RUNNING", "PAUSED", "STOPPED"
	Progress     float64       `json:"progress"` // 0.0 - 1.0
	SessionCount int           `json:"session_count"`
}

// PomodoroEventData contiene datos específicos de eventos de pomodoro
type PomodoroEventData struct {
	Number       int           `json:"number"`
	Duration     time.Duration `json:"duration"`
	ActualTime   time.Duration `json:"actual_time"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	NextBreak    string        `json:"next_break"`
	NextDuration time.Duration `json:"next_duration"`
}

// BreakEventData contiene datos específicos de eventos de break
type BreakEventData struct {
	Type        string        `json:"type"` // "DESCANSO", "DESCANSO LARGO"
	Duration    time.Duration `json:"duration"`
	ActualTime  time.Duration `json:"actual_time"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	IsLongBreak bool          `json:"is_long_break"`
}

// StatsEventData contiene datos específicos de eventos de estadísticas
type StatsEventData struct {
	PomodorosCompleted int           `json:"pomodoros_completed"`
	PomodorosSkipped   int           `json:"pomodoros_skipped"`
	BreaksCompleted    int           `json:"breaks_completed"`
	BreaksSkipped      int           `json:"breaks_skipped"`
	CurrentStreak      int           `json:"current_streak"`
	BestStreak         int           `json:"best_streak"`
	TotalWorkTime      time.Duration `json:"total_work_time"`
	TotalBreakTime     time.Duration `json:"total_break_time"`
	SessionDuration    time.Duration `json:"session_duration"`
	WorkEfficiency     float64       `json:"work_efficiency"`
}

// SessionEventData contiene datos específicos de eventos de sesión
type SessionEventData struct {
	SessionID  string        `json:"session_id"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	TotalTime  time.Duration `json:"total_time"`
	ConfigUsed interface{}   `json:"config_used"`
}

// ErrorEventData contiene datos específicos de eventos de error
type ErrorEventData struct {
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Source  string      `json:"source"`
	Details interface{} `json:"details,omitempty"`
}

// Helper functions para crear eventos comunes

// NewTimerEvent crea un evento de timer con los datos proporcionados
func NewTimerEvent(eventType EventType, remaining, total time.Duration, state, status string, progress float64, sessionCount int) Event {
	return Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data: TimerEventData{
			Remaining:    remaining,
			Total:        total,
			State:        state,
			Status:       status,
			Progress:     progress,
			SessionCount: sessionCount,
		},
	}
}

// NewErrorEvent crea un evento de error
func NewErrorEvent(message, code, source string, details interface{}) Event {
	return Event{
		Type:      ErrorOccurred,
		Timestamp: time.Now(),
		Data: ErrorEventData{
			Message: message,
			Code:    code,
			Source:  source,
			Details: details,
		},
	}
}
