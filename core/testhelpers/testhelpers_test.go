package testhelpers

import (
	"testing"
	"time"

	"github.com/kubaliski/pomodoro-core/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupVerification verifica que el setup de testing funciona correctamente
func TestSetupVerification(t *testing.T) {
	t.Run("testify works", func(t *testing.T) {
		assert.True(t, true)
		require.NotNil(t, t)
	})

	t.Run("fast config is valid", func(t *testing.T) {
		cfg := FastTestConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, 1*time.Minute, cfg.WorkDuration)
		assert.NoError(t, cfg.Validate())
	})

	t.Run("medium config is valid", func(t *testing.T) {
		cfg := MediumTestConfig()
		require.NotNil(t, cfg)
		assert.Equal(t, 3*time.Minute, cfg.WorkDuration)
		assert.NoError(t, cfg.Validate())
	})
}

// TestEventCollector verifica que el EventCollector funciona correctamente
func TestEventCollector(t *testing.T) {
	collector := NewEventCollector()
	require.NotNil(t, collector)

	t.Run("initial state", func(t *testing.T) {
		eventsList := collector.GetEvents()
		assert.Empty(t, eventsList)

		types := collector.GetEventTypes()
		assert.Empty(t, types)

		count := collector.CountEventsByType(events.TimerTick)
		assert.Equal(t, 0, count)
	})

	t.Run("collect events", func(t *testing.T) {
		// Simular eventos
		event1 := events.Event{
			Type:      events.TimerStarted,
			Timestamp: time.Now(),
			Data:      "test data 1",
		}

		event2 := events.Event{
			Type:      events.TimerTick,
			Timestamp: time.Now(),
			Data:      "test data 2",
		}

		collector.HandleEvent(event1)
		collector.HandleEvent(event2)

		// Verificar que se recopilaron
		allEvents := collector.GetEvents()
		assert.Len(t, allEvents, 2)
		assert.Equal(t, events.TimerStarted, allEvents[0].Type)
		assert.Equal(t, events.TimerTick, allEvents[1].Type)

		// Verificar tipos
		types := collector.GetEventTypes()
		assert.Len(t, types, 2)
		assert.Equal(t, events.TimerStarted, types[0])
		assert.Equal(t, events.TimerTick, types[1])

		// Verificar filtrado por tipo
		timerStartedEvents := collector.GetEventsByType(events.TimerStarted)
		assert.Len(t, timerStartedEvents, 1)
		assert.Equal(t, "test data 1", timerStartedEvents[0].Data)

		timerTickEvents := collector.GetEventsByType(events.TimerTick)
		assert.Len(t, timerTickEvents, 1)
		assert.Equal(t, "test data 2", timerTickEvents[0].Data)

		// Verificar conteo
		assert.Equal(t, 1, collector.CountEventsByType(events.TimerStarted))
		assert.Equal(t, 1, collector.CountEventsByType(events.TimerTick))
		assert.Equal(t, 0, collector.CountEventsByType(events.TimerCompleted))
	})

	t.Run("clear events", func(t *testing.T) {
		collector.Clear()

		eventsList := collector.GetEvents()
		assert.Empty(t, eventsList)

		count := collector.CountEventsByType(events.TimerStarted)
		assert.Equal(t, 0, count)
	})
}

// TestMockEventHandler verifica que el MockEventHandler funciona correctamente
func TestMockEventHandler(t *testing.T) {
	mock := NewMockEventHandler()
	require.NotNil(t, mock)

	t.Run("initial state", func(t *testing.T) {
		assert.Equal(t, 0, mock.GetHandleCalls())
		assert.Nil(t, mock.GetLastEvent())
		assert.Empty(t, mock.GetAllEvents())
	})

	t.Run("handle events", func(t *testing.T) {
		event := events.Event{
			Type:      events.TimerStarted,
			Timestamp: time.Now(),
			Data:      "test data",
		}

		mock.HandleEvent(event)

		assert.Equal(t, 1, mock.GetHandleCalls())
		assert.NotNil(t, mock.GetLastEvent())
		assert.Equal(t, events.TimerStarted, mock.GetLastEvent().Type)

		allEvents := mock.GetAllEvents()
		assert.Len(t, allEvents, 1)
		assert.Equal(t, events.TimerStarted, allEvents[0].Type)
	})

	t.Run("reset mock", func(t *testing.T) {
		mock.Reset()

		assert.Equal(t, 0, mock.GetHandleCalls())
		assert.Nil(t, mock.GetLastEvent())
		assert.Empty(t, mock.GetAllEvents())
	})
}
