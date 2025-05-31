package handlers

import (
	"fmt"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/ui"
	"github.com/kubaliski/pomodoro-core/events"
)

// EventHandler maneja todos los eventos del core del pomodoro
type EventHandler struct {
	handler *CLIHandler
}

// NewEventHandler crea un nuevo manejador de eventos
func NewEventHandler(h *CLIHandler) *EventHandler {
	return &EventHandler{handler: h}
}

// SetupEventHandlers configura todos los manejadores de eventos
func (eh *EventHandler) SetupEventHandlers(eventBus *events.EventBus) {
	// Timer events
	eventBus.SubscribeFunc(events.TimerStarted, eh.HandleTimerStarted)
	eventBus.SubscribeFunc(events.TimerTick, eh.HandleTimerTick)
	eventBus.SubscribeFunc(events.TimerPaused, eh.HandleTimerPaused)
	eventBus.SubscribeFunc(events.TimerResumed, eh.HandleTimerResumed)
	eventBus.SubscribeFunc(events.TimerCompleted, eh.HandleTimerCompleted)
	eventBus.SubscribeFunc(events.TimerSkipped, eh.HandleTimerSkipped)

	// Session events
	eventBus.SubscribeFunc(events.PomodoroStarted, eh.HandlePomodoroStarted)
	eventBus.SubscribeFunc(events.PomodoroCompleted, eh.HandlePomodoroCompleted)
	eventBus.SubscribeFunc(events.PomodoroSkipped, eh.HandlePomodoroSkipped)
	eventBus.SubscribeFunc(events.BreakStarted, eh.HandleBreakStarted)
	eventBus.SubscribeFunc(events.BreakCompleted, eh.HandleBreakCompleted)
	eventBus.SubscribeFunc(events.BreakSkipped, eh.HandleBreakSkipped)

	// Stats events
	eventBus.SubscribeFunc(events.StatsUpdated, eh.HandleStatsUpdated)

	// Engine events
	eventBus.SubscribeFunc(events.EngineStarted, eh.HandleEngineStarted)
	eventBus.SubscribeFunc(events.EngineStopped, eh.HandleEngineStopped)
}

// Timer Event Handlers

func (eh *EventHandler) HandleEngineStarted(event events.Event) {
	// El engine ha iniciado pero aún no hay sesión corriendo
}

func (eh *EventHandler) HandleEngineStopped(event events.Event) {
	fmt.Println("🛑 Engine detenido.")
}

func (eh *EventHandler) HandleTimerStarted(event events.Event) {
	if data, ok := event.Data.(events.TimerEventData); ok {
		eh.handler.SetCurrentTimerData(data)
		eh.handler.SetFirstSessionStarted(true)
		eh.handler.UpdateLastAlert(-1) // Reset alert tracking

		// 🔊 Notificación de inicio de sesión
		sessionType := data.State
		duration := time.Duration(data.Total) * time.Nanosecond
		eh.handler.GetNotificationManager().NotifySessionStarted(sessionType, duration)

		// Limpiar línea de comando y mostrar display inicial
		fmt.Print("\r\033[K")
		eh.handler.GetUIHelpers().DisplayTimerWithStats()
		fmt.Println()
		fmt.Print("Comando > ")
	}
}

func (eh *EventHandler) HandleTimerTick(event events.Event) {
	if data, ok := event.Data.(events.TimerEventData); ok {
		eh.handler.SetCurrentTimerData(data)

		showing := eh.handler.IsShowingStats() || eh.handler.IsWaitingForInput()

		// 🔊 ALERTAS DE TIEMPO INTELIGENTES
		eh.handleTimeAlerts(data)

		// Solo actualizar si no estamos mostrando mensajes importantes
		if !showing {
			// Actualizar display sin interrumpir input
			fmt.Print("\033[s")   // Guardar cursor
			fmt.Print("\033[A")   // Subir una línea
			fmt.Print("\r\033[K") // Limpiar línea del timer
			eh.handler.GetUIHelpers().DisplayTimerWithStats()
			fmt.Print("\033[u") // Restaurar cursor
		}
	}
}

func (eh *EventHandler) handleTimeAlerts(data events.TimerEventData) {
	timeRemaining := time.Duration(data.Remaining) * time.Nanosecond
	sessionType := data.State
	currentMinute := int(timeRemaining.Minutes())
	lastAlertMinute := eh.handler.GetLastAlertMinute()

	// Evitar spam de alertas - solo alertar cuando cambia el minuto
	if currentMinute == lastAlertMinute {
		return
	}

	shouldAlert := false

	// Alertas en minutos específicos: 5, 2, 1
	switch currentMinute {
	case 5, 2, 1:
		if timeRemaining.Seconds()-float64(currentMinute*60) < 5 { // Solo en los primeros 5 segundos del minuto
			shouldAlert = true
		}
	}

	// Alerta especial a los 30 segundos
	if timeRemaining <= 30*time.Second && timeRemaining > 25*time.Second && lastAlertMinute != 0 {
		shouldAlert = true
		currentMinute = 0 // Marcar como alerta de 30 segundos
	}

	if shouldAlert {
		eh.handler.UpdateLastAlert(currentMinute)
		eh.handler.GetNotificationManager().NotifyTimeAlert(timeRemaining, sessionType)
	}
}

func (eh *EventHandler) HandleTimerPaused(event events.Event) {
	// 🔊 Notificación de pausa
	eh.handler.GetNotificationManager().NotifyTimerPaused()

	fmt.Println("⏸️  Timer pausado. Escribe 'r' para reanudar.")
	fmt.Print("Comando > ")
}

func (eh *EventHandler) HandleTimerResumed(event events.Event) {
	// 🔊 Notificación de reanudación
	timeRemaining := time.Duration(eh.handler.GetCurrentTimerData().Remaining) * time.Nanosecond
	eh.handler.GetNotificationManager().NotifyTimerResumed(timeRemaining)

	fmt.Println("▶️  Timer reanudado.")
	fmt.Print("Comando > ")
}

func (eh *EventHandler) HandleTimerCompleted(event events.Event) {
	fmt.Println() // Nueva línea al terminar
}

func (eh *EventHandler) HandleTimerSkipped(event events.Event) {
	fmt.Println("⏭️  Timer saltado.")
}

// Session Event Handlers

func (eh *EventHandler) HandlePomodoroStarted(event events.Event) {
	if data, ok := event.Data.(events.PomodoroEventData); ok {
		fmt.Printf("\n🍅 Pomodoro #%d - Sesión de trabajo\n", data.Number)
		time.Sleep(2 * time.Second)
	}
}

func (eh *EventHandler) HandlePomodoroCompleted(event events.Event) {
	if data, ok := event.Data.(events.PomodoroEventData); ok {
		eh.handler.SetWaitingForInput(true)

		// 🔊 NOTIFICACIÓN DE POMODORO COMPLETADO
		nextBreakType, nextDuration := eh.handler.GetNextBreakInfo(data.Number)
		eh.handler.GetNotificationManager().NotifyPomodoroCompleted(data.Number, nextDuration)

		// Limpiar display antes de mostrar mensaje
		fmt.Print("\r\033[K") // Limpiar línea actual
		fmt.Println()
		fmt.Println()

		fmt.Println(ui.Colorize("+================================+", ui.ColorGreen, true))
		fmt.Println(ui.Colorize("|       POMODORO COMPLETO!       |", ui.ColorGreen, true))
		fmt.Println(ui.Colorize("+================================+", ui.ColorGreen, true))
		fmt.Printf("✅ ¡Pomodoro #%d completado!\n", data.Number)

		// Determinar próximo descanso CORRECTAMENTE
		fmt.Printf("🎯 Próximo: %s (%s)\n", nextBreakType, ui.FormatDuration(nextDuration))
		fmt.Println()

		// Mostrar estadísticas rápidas
		stats := eh.handler.GetCurrentStatsData()
		fmt.Printf("📊 Estadísticas: 🍅 %d | 🔥 %d | ⏱️ %s\n",
			stats.PomodorosCompleted, stats.CurrentStreak, FormatDuration(stats.TotalWorkTime))

		if stats.CurrentStreak > 1 {
			fmt.Printf("🔥 ¡Racha de %d pomodoros!\n", stats.CurrentStreak)
		}
		fmt.Println()

		fmt.Println(ui.Colorize("Escribe 'c' para continuar, 'stats' para ver estadísticas detalladas, o 'q' para salir", ui.ColorYellow, true))
		fmt.Print("Comando > ")

		eh.handler.SetWaitingForInput(false)
	}
}

func (eh *EventHandler) HandlePomodoroSkipped(event events.Event) {
	if data, ok := event.Data.(events.PomodoroEventData); ok {
		eh.handler.SetWaitingForInput(true)

		// Limpiar display antes de mostrar mensaje
		fmt.Print("\r\033[K") // Limpiar línea actual
		fmt.Println()
		fmt.Println()

		fmt.Println(ui.Colorize("+================================+", ui.ColorRed, true))
		fmt.Println(ui.Colorize("|      POMODORO SALTADO!         |", ui.ColorRed, true))
		fmt.Println(ui.Colorize("+================================+", ui.ColorRed, true))

		// Mostrar número correcto del pomodoro
		pomodoroNum := data.Number
		if pomodoroNum == 0 {
			pomodoroNum = eh.handler.GetEngine().GetPomodoroCount() + 1
		}
		fmt.Printf("⏭️  Pomodoro #%d saltado\n", pomodoroNum)

		// Determinar próximo descanso CORRECTAMENTE
		nextBreakType, nextDuration := eh.handler.GetNextBreakInfo(pomodoroNum)
		fmt.Printf("🎯 Próximo: %s (%s)\n", nextBreakType, ui.FormatDuration(nextDuration))
		fmt.Println()

		// Mensaje claro de continuación
		fmt.Println(ui.Colorize("Escribe 'c' para continuar con el descanso o 'q' para salir", ui.ColorYellow, true))
		fmt.Print("Comando > ")

		eh.handler.SetWaitingForInput(false)
	}
}

func (eh *EventHandler) HandleBreakStarted(event events.Event) {
	if data, ok := event.Data.(events.BreakEventData); ok {
		fmt.Printf("\n🧘 %s - Tiempo de descanso\n", data.Type)
		time.Sleep(2 * time.Second)
	}
}

func (eh *EventHandler) HandleBreakCompleted(event events.Event) {
	if data, ok := event.Data.(events.BreakEventData); ok {
		eh.handler.SetWaitingForInput(true)

		// 🔊 NOTIFICACIÓN DE DESCANSO COMPLETADO
		nextPomodoroNum := eh.handler.GetEngine().GetPomodoroCount() + 1
		eh.handler.GetNotificationManager().NotifyBreakCompleted(data.Type, nextPomodoroNum)

		// Limpiar display antes de mostrar mensaje
		fmt.Print("\r\033[K") // Limpiar línea actual
		fmt.Println()
		fmt.Println()

		fmt.Println(ui.Colorize("+================================+", ui.ColorBlue, true))
		fmt.Println(ui.Colorize("|      DESCANSO COMPLETADO!      |", ui.ColorBlue, true))
		fmt.Println(ui.Colorize("+================================+", ui.ColorBlue, true))
		fmt.Printf("✅ %s terminado\n", data.Type)
		fmt.Println("💪 ¡Listo para el siguiente pomodoro!")
		fmt.Println()

		// Mostrar qué pomodoro viene
		fmt.Printf("🎯 Próximo: Pomodoro #%d (%s)\n",
			nextPomodoroNum, ui.FormatDuration(eh.handler.GetEngine().GetConfig().WorkDuration))
		fmt.Println()

		fmt.Println(ui.Colorize("Escribe 'c' para continuar o 'q' para salir", ui.ColorYellow, true))
		fmt.Print("Comando > ")

		eh.handler.SetWaitingForInput(false)
	}
}

func (eh *EventHandler) HandleBreakSkipped(event events.Event) {
	if data, ok := event.Data.(events.BreakEventData); ok {
		eh.handler.SetWaitingForInput(true)

		// Limpiar display antes de mostrar mensaje
		fmt.Print("\r\033[K") // Limpiar línea actual
		fmt.Println()
		fmt.Println()

		fmt.Println(ui.Colorize("+================================+", ui.ColorCyan, true))
		fmt.Println(ui.Colorize("|      DESCANSO SALTADO!         |", ui.ColorCyan, true))
		fmt.Println(ui.Colorize("+================================+", ui.ColorCyan, true))
		fmt.Printf("⏭️  %s saltado\n", data.Type)
		fmt.Println("💪 ¡Listo para el siguiente pomodoro!")
		fmt.Println()

		// Mostrar qué pomodoro viene
		nextPomodoroNum := eh.handler.GetEngine().GetPomodoroCount() + 1
		fmt.Printf("🎯 Próximo: Pomodoro #%d (%s)\n",
			nextPomodoroNum, ui.FormatDuration(eh.handler.GetEngine().GetConfig().WorkDuration))
		fmt.Println()

		fmt.Println(ui.Colorize("Escribe 'c' para continuar con el trabajo o 'q' para salir", ui.ColorYellow, true))
		fmt.Print("Comando > ")

		eh.handler.SetWaitingForInput(false)
	}
}

func (eh *EventHandler) HandleStatsUpdated(event events.Event) {
	if data, ok := event.Data.(events.StatsEventData); ok {
		eh.handler.SetCurrentStatsData(data)
	}
}
