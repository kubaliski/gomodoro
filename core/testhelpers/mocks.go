package testhelpers

import (
	"sync"
	"time"

	"github.com/kubaliski/pomodoro-core/events"
)

// MockEventHandler es un mock de EventHandler para testing
type MockEventHandler struct {
	mu           sync.RWMutex
	handleCalls  int
	lastEvent    *events.Event
	allEvents    []events.Event
	shouldPanic  bool
	panicMessage string
}

// NewMockEventHandler crea un nuevo mock handler
func NewMockEventHandler() *MockEventHandler {
	return &MockEventHandler{
		allEvents: make([]events.Event, 0),
	}
}

// HandleEvent implementa events.EventHandler
func (m *MockEventHandler) HandleEvent(event events.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldPanic {
		panic(m.panicMessage)
	}

	m.handleCalls++
	m.lastEvent = &event
	m.allEvents = append(m.allEvents, event)
}

// GetHandleCalls retorna el número de veces que se llamó HandleEvent
func (m *MockEventHandler) GetHandleCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.handleCalls
}

// GetLastEvent retorna el último evento recibido
func (m *MockEventHandler) GetLastEvent() *events.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastEvent
}

// GetAllEvents retorna todos los eventos recibidos
func (m *MockEventHandler) GetAllEvents() []events.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()

	eventsCopy := make([]events.Event, len(m.allEvents))
	copy(eventsCopy, m.allEvents)
	return eventsCopy
}

// SetShouldPanic configura el mock para hacer panic en HandleEvent
func (m *MockEventHandler) SetShouldPanic(shouldPanic bool, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldPanic = shouldPanic
	m.panicMessage = message
}

// Reset limpia el estado del mock
func (m *MockEventHandler) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handleCalls = 0
	m.lastEvent = nil
	m.allEvents = make([]events.Event, 0)
	m.shouldPanic = false
	m.panicMessage = ""
}

// MockTimeProvider simula el tiempo para tests determinísticos
type MockTimeProvider struct {
	mu          sync.RWMutex
	currentTime time.Time
	frozen      bool
}

// NewMockTimeProvider crea un nuevo proveedor de tiempo mock
func NewMockTimeProvider(startTime time.Time) *MockTimeProvider {
	return &MockTimeProvider{
		currentTime: startTime,
		frozen:      false,
	}
}

// Now retorna el tiempo actual del mock
func (m *MockTimeProvider) Now() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTime
}

// Advance avanza el tiempo del mock
func (m *MockTimeProvider) Advance(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.frozen {
		m.currentTime = m.currentTime.Add(duration)
	}
}

// SetTime establece un tiempo específico
func (m *MockTimeProvider) SetTime(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentTime = t
}

// Freeze congela el tiempo (no avanza automáticamente)
func (m *MockTimeProvider) Freeze() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.frozen = true
}

// Unfreeze permite que el tiempo avance nuevamente
func (m *MockTimeProvider) Unfreeze() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.frozen = false
}

// IsFrozen retorna si el tiempo está congelado
func (m *MockTimeProvider) IsFrozen() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.frozen
}

// EventCapture captura eventos con filtros para análisis detallado
type EventCapture struct {
	mu       sync.RWMutex
	events   []events.Event
	filters  map[events.EventType]bool
	maxSize  int
	overflow bool
}

// NewEventCapture crea un nuevo capturador de eventos
func NewEventCapture(maxSize int) *EventCapture {
	return &EventCapture{
		events:  make([]events.Event, 0),
		filters: make(map[events.EventType]bool),
		maxSize: maxSize,
	}
}

// HandleEvent implementa events.EventHandler con filtros
func (ec *EventCapture) HandleEvent(event events.Event) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	// Verificar filtros
	if len(ec.filters) > 0 {
		if !ec.filters[event.Type] {
			return // Evento filtrado
		}
	}

	// Verificar límite de tamaño
	if len(ec.events) >= ec.maxSize {
		ec.overflow = true
		return
	}

	ec.events = append(ec.events, event)
}

// AddFilter agrega un filtro para capturar solo ciertos tipos de eventos
func (ec *EventCapture) AddFilter(eventType events.EventType) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.filters[eventType] = true
}

// RemoveFilter remueve un filtro
func (ec *EventCapture) RemoveFilter(eventType events.EventType) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	delete(ec.filters, eventType)
}

// ClearFilters remueve todos los filtros
func (ec *EventCapture) ClearFilters() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.filters = make(map[events.EventType]bool)
}

// GetCapturedEvents retorna una copia de los eventos capturados
func (ec *EventCapture) GetCapturedEvents() []events.Event {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	eventsCopy := make([]events.Event, len(ec.events))
	copy(eventsCopy, ec.events)
	return eventsCopy
}

// HasOverflow indica si se perdieron eventos por límite de tamaño
func (ec *EventCapture) HasOverflow() bool {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.overflow
}

// GetCaptureStats retorna estadísticas de captura
func (ec *EventCapture) GetCaptureStats() (captured int, overflow bool, filters int) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return len(ec.events), ec.overflow, len(ec.filters)
}

// Reset limpia todos los eventos capturados
func (ec *EventCapture) Reset() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.events = make([]events.Event, 0)
	ec.overflow = false
}
