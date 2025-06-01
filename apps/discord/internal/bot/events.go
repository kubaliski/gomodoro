package bot

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kubaliski/gomodoro/apps/discord/internal/manager"
	"github.com/kubaliski/pomodoro-core/config"
	"github.com/kubaliski/pomodoro-core/events"
)

// EventHandler maneja los eventos de pomodoro y los convierte en notificaciones de Discord
type EventHandler struct {
	sessionManager *manager.SessionManager
}

// NewEventHandler crea una nueva instancia del event handler
func NewEventHandler(sessionManager *manager.SessionManager) *EventHandler {
	return &EventHandler{
		sessionManager: sessionManager,
	}
}

// RegisterWithSessionManager registra todos los event handlers con el session manager
func (eh *EventHandler) RegisterWithSessionManager(notifier *NotificationManager) {
	log.Printf("🔧 Registering pomodoro event handlers...")

	eh.sessionManager.RegisterEventHandler("pomodoro_completed", eh.createPomodoroCompletedHandler(notifier))
	eh.sessionManager.RegisterEventHandler("break_completed", eh.createBreakCompletedHandler(notifier))
	eh.sessionManager.RegisterEventHandler("pomodoro_started", eh.createPomodoroStartedHandler(notifier))
	eh.sessionManager.RegisterEventHandler("break_started", eh.createBreakStartedHandler(notifier))
	eh.sessionManager.RegisterEventHandler("timer_reminder", eh.createTimerReminderHandler(notifier))

	log.Printf("✅ All event handlers registered successfully")
}

// createPomodoroCompletedHandler crea el handler para cuando se completa un pomodoro
func (eh *EventHandler) createPomodoroCompletedHandler(notifier *NotificationManager) manager.EventHandlerFunc {
	return func(userID, channelID string, event events.Event) {
		data, ok := event.Data.(events.PomodoroEventData)
		if !ok {
			log.Printf("❌ Invalid event data type for PomodoroCompleted")
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       "🎉 ¡Pomodoro Completado!",
			Description: fmt.Sprintf("¡Excelente trabajo! Has completado el pomodoro #%d", data.Number),
			Color:       0x00ff00,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Duración Configurada", Value: config.FormatDuration(data.Duration), Inline: true},
				{Name: "Tiempo Real", Value: config.FormatDuration(data.ActualTime), Inline: true},
				{Name: "Eficiencia", Value: fmt.Sprintf("%.1f%%", eh.calculateEfficiency(data.Duration, data.ActualTime)), Inline: true},
			},
			Timestamp: time.Now().Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "¡Momento perfecto para un descanso merecido!",
			},
		}

		mention := "¡Hora de un descanso! 🧘‍♂️"

		if err := notifier.SendNotification(userID, channelID, embed, mention); err != nil {
			log.Printf("❌ Error sending pomodoro completed notification: %v", err)
		}
	}
}

// createBreakCompletedHandler crea el handler para cuando se completa un descanso
func (eh *EventHandler) createBreakCompletedHandler(notifier *NotificationManager) manager.EventHandlerFunc {
	return func(userID, channelID string, event events.Event) {
		data, ok := event.Data.(events.BreakEventData)
		if !ok {
			log.Printf("❌ Invalid event data type for BreakCompleted")
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       "⏰ ¡Descanso Completado!",
			Description: "El tiempo de descanso ha terminado. ¿Listo para volver al trabajo?",
			Color:       0xffa500,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Tipo de Descanso", Value: translateBreakType(data.Type), Inline: true},
				{Name: "Duración", Value: config.FormatDuration(data.ActualTime), Inline: true},
			},
			Timestamp: time.Now().Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "¡A concentrarse en la siguiente sesión!",
			},
		}

		mention := "¡De vuelta al trabajo! 💪"

		if err := notifier.SendNotification(userID, channelID, embed, mention); err != nil {
			log.Printf("❌ Error sending break completed notification: %v", err)
		}
	}
}

// createPomodoroStartedHandler crea el handler para cuando inicia un pomodoro
func (eh *EventHandler) createPomodoroStartedHandler(notifier *NotificationManager) manager.EventHandlerFunc {
	return func(userID, channelID string, event events.Event) {
		data, ok := event.Data.(events.PomodoroEventData)
		if !ok {
			log.Printf("❌ Invalid event data type for PomodoroStarted")
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       "🍅 ¡Hora de Concentrarse!",
			Description: fmt.Sprintf("Pomodoro #%d iniciado - ¡hora de enfocarse en tu trabajo!", data.Number),
			Color:       0xff6b6b,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Duración", Value: config.FormatDuration(data.Duration), Inline: true},
				{Name: "Iniciado", Value: data.StartTime.Format("15:04:05"), Inline: true},
			},
			Timestamp: time.Now().Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Elimina las distracciones y concéntrate",
			},
		}

		if err := notifier.SendNotification(userID, channelID, embed, ""); err != nil {
			log.Printf("❌ Error sending pomodoro started notification: %v", err)
		}
	}
}

// createBreakStartedHandler crea el handler para cuando inicia un descanso
func (eh *EventHandler) createBreakStartedHandler(notifier *NotificationManager) manager.EventHandlerFunc {
	return func(userID, channelID string, event events.Event) {
		data, ok := event.Data.(events.BreakEventData)
		if !ok {
			log.Printf("❌ Invalid event data type for BreakStarted")
			return
		}

		breakType := "Descanso Corto"
		emoji := "☕"
		tip := "Levántate, estírate o toma algo de agua"

		if data.IsLongBreak {
			breakType = "Descanso Largo"
			emoji = "🏖️"
			tip = "Tiempo perfecto para una caminata o una comida"
		}

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s %s Iniciado", emoji, breakType),
			Description: fmt.Sprintf("Hora de relajarse por %s", config.FormatDuration(data.Duration)),
			Color:       0x0099ff,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "💡 Sugerencia", Value: tip, Inline: false},
			},
			Timestamp: time.Now().Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Un buen descanso mejora la productividad",
			},
		}

		if err := notifier.SendNotification(userID, channelID, embed, ""); err != nil {
			log.Printf("❌ Error sending break started notification: %v", err)
		}
	}
}

// createTimerReminderHandler crea el handler para recordatorios de tiempo
func (eh *EventHandler) createTimerReminderHandler(notifier *NotificationManager) manager.EventHandlerFunc {
	return func(userID, channelID string, event events.Event) {
		data, ok := event.Data.(events.TimerEventData)
		if !ok {
			log.Printf("❌ Invalid event data type for TimerReminder")
			return
		}

		remaining := int(data.Remaining.Minutes())

		var message string
		var color int
		var emoji string

		switch remaining {
		case 10:
			message = "Quedan 10 minutos"
			color = 0xffaa00
			emoji = "⏰"
		case 5:
			message = "Quedan 5 minutos"
			color = 0xff6600
			emoji = "⏰"
		case 1:
			message = "¡Queda 1 minuto!"
			color = 0xff0000
			emoji = "🚨"
		default:
			return // No reminder needed for other times
		}

		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s Recordatorio de Tiempo", emoji),
			Description: message,
			Color:       color,
			Timestamp:   time.Now().Format(time.RFC3339),
		}

		if err := notifier.SendNotification(userID, channelID, embed, ""); err != nil {
			log.Printf("❌ Error sending timer reminder: %v", err)
		}
	}
}

// calculateEfficiency calcula la eficiencia basada en tiempo configurado vs tiempo real
func (eh *EventHandler) calculateEfficiency(planned, actual time.Duration) float64 {
	if planned == 0 {
		return 0
	}

	// Si el tiempo real es menor o igual al planeado, eficiencia alta
	if actual <= planned {
		return 100.0
	}

	// Si se tardó más, calcular porcentaje basado en tiempo extra
	efficiency := float64(planned) / float64(actual) * 100
	if efficiency < 0 {
		efficiency = 0
	}

	return efficiency
}
