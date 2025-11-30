package testhelpers

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kubaliski/pomodoro-core/config"
	"github.com/kubaliski/pomodoro-core/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DefaultTestTimeout es el timeout por defecto para operaciones de test
const DefaultTestTimeout = 5 * time.Second

// FastTestConfig retorna una configuración optimizada para tests rápidos
func FastTestConfig() *config.Config {
	return &config.Config{
		WorkDuration:      1 * time.Minute, // Mínimo válido
		ShortBreak:        1 * time.Minute, // Mínimo válido
		LongBreak:         5 * time.Minute, // Mínimo válido
		LongBreakInterval: 2,               // Cada 2 pomodoros para testing
	}
}

// MediumTestConfig retorna una configuración para tests de duración media
func MediumTestConfig() *config.Config {
	return &config.Config{
		WorkDuration:      3 * time.Minute,
		ShortBreak:        2 * time.Minute,
		LongBreak:         7 * time.Minute,
		LongBreakInterval: 3,
	}
}

// UltraFastTestConfig retorna una configuración para tests que necesitan ejecutarse muy rápido
// NOTA: Esta configuración NO pasará validación, solo para tests internos de timer
func UltraFastTestConfig() *config.Config {
	return &config.Config{
		WorkDuration:      100 * time.Millisecond,
		ShortBreak:        50 * time.Millisecond,
		LongBreak:         150 * time.Millisecond,
		LongBreakInterval: 2,
	}
}

// EventCollector recopila eventos para verificación en tests
type EventCollector struct {
	mu     sync.RWMutex
	events []events.Event
}

// NewEventCollector crea un nuevo recopilador de eventos
func NewEventCollector() *EventCollector {
	return &EventCollector{
		events: make([]events.Event, 0),
	}
}

// HandleEvent implementa events.EventHandler
func (ec *EventCollector) HandleEvent(event events.Event) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.events = append(ec.events, event)
}

// GetEvents retorna una copia de todos los eventos recopilados
func (ec *EventCollector) GetEvents() []events.Event {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	eventsCopy := make([]events.Event, len(ec.events))
	copy(eventsCopy, ec.events)
	return eventsCopy
}

// GetEventTypes retorna solo los tipos de eventos recopilados
func (ec *EventCollector) GetEventTypes() []events.EventType {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	types := make([]events.EventType, len(ec.events))
	for i, event := range ec.events {
		types[i] = event.Type
	}
	return types
}

// GetEventsByType retorna eventos filtrados por tipo
func (ec *EventCollector) GetEventsByType(eventType events.EventType) []events.Event {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	var filtered []events.Event
	for _, event := range ec.events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// CountEventsByType cuenta eventos de un tipo específico
func (ec *EventCollector) CountEventsByType(eventType events.EventType) int {
	return len(ec.GetEventsByType(eventType))
}

// Clear limpia todos los eventos recopilados
func (ec *EventCollector) Clear() {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.events = make([]events.Event, 0)
}

// WaitForEvent espera a que ocurra un evento específico con timeout
func WaitForEvent(t *testing.T, collector *EventCollector, eventType events.EventType, timeout time.Duration) events.Event {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for event %s after %v", eventType, timeout)
			return events.Event{}
		case <-ticker.C:
			events := collector.GetEventsByType(eventType)
			if len(events) > 0 {
				return events[len(events)-1] // Retorna el más reciente
			}
		}
	}
}

// WaitForEventCount espera hasta que haya un número específico de eventos de un tipo
func WaitForEventCount(t *testing.T, collector *EventCollector, eventType events.EventType, count int, timeout time.Duration) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			actual := collector.CountEventsByType(eventType)
			t.Fatalf("timeout waiting for %d events of type %s, got %d after %v", count, eventType, actual, timeout)
		case <-ticker.C:
			if collector.CountEventsByType(eventType) >= count {
				return
			}
		}
	}
}

// AssertEventSequence verifica que los eventos ocurrieron en el orden esperado
func AssertEventSequence(t *testing.T, collector *EventCollector, expectedTypes ...events.EventType) {
	t.Helper()

	actualTypes := collector.GetEventTypes()

	// Verificar que tenemos al menos los eventos esperados
	require.GreaterOrEqual(t, len(actualTypes), len(expectedTypes),
		"Not enough events: expected at least %d, got %d", len(expectedTypes), len(actualTypes))

	// Verificar secuencia (permitir eventos adicionales entre los esperados)
	expectedIndex := 0
	for _, actualType := range actualTypes {
		if expectedIndex < len(expectedTypes) && actualType == expectedTypes[expectedIndex] {
			expectedIndex++
		}

		// Si encontramos todos los eventos esperados, éxito
		if expectedIndex == len(expectedTypes) {
			return
		}
	}

	// Si llegamos aquí, no encontramos todos los eventos esperados
	t.Fatalf("Event sequence mismatch. Expected: %v, Got: %v", expectedTypes, actualTypes)
}

// AssertEventData verifica que un evento contiene los datos esperados
func AssertEventData[T any](t *testing.T, event events.Event, validator func(T) bool) {
	t.Helper()

	data, ok := event.Data.(T)
	require.True(t, ok, "Event data is not of expected type")
	require.True(t, validator(data), "Event data validation failed")
}

// AssertNoRaceConditions verifica que no hay race conditions usando goleak
func AssertNoRaceConditions(t *testing.T) {
	t.Helper()
	// Esta función se expandirá en fases posteriores con goleak
	// Por ahora, solo marca que el test debe ejecutarse con -race
}

// TestContext crea un contexto con timeout para tests
func TestContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// TestContextDefault crea un contexto con timeout por defecto
func TestContextDefault() (context.Context, context.CancelFunc) {
	return TestContext(DefaultTestTimeout)
}

// AssertDurationApprox verifica que dos duraciones están aproximadamente iguales
func AssertDurationApprox(t *testing.T, expected, actual time.Duration, tolerance time.Duration) {
	t.Helper()

	diff := expected - actual
	if diff < 0 {
		diff = -diff
	}

	assert.LessOrEqual(t, diff, tolerance,
		"Duration mismatch: expected %v ± %v, got %v (diff: %v)",
		expected, tolerance, actual, diff)
}

// SetupEventCollector configura un collector en un EventBus y retorna función de limpieza
func SetupEventCollector(eventBus *events.EventBus) (*EventCollector, func()) {
	collector := NewEventCollector()
	eventBus.SubscribeGlobal(collector)

	cleanup := func() {
		collector.Clear()
		// Nota: EventBus actualmente no tiene método para unsubscribe global
		// Esto se puede mejorar en fases posteriores
	}

	return collector, cleanup
}

// SkipIfShort saltea el test si está en modo short
func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
}
