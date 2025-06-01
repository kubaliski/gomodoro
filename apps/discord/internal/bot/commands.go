package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kubaliski/pomodoro-core/config"
	"github.com/kubaliski/pomodoro-core/stats"
)

// handlePausePomodoro maneja el comando de pausar pomodoro
func (b *Bot) handlePausePomodoro(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID, err := b.getUserID(i)
	if err != nil {
		b.respondWithError(s, i, err.Error())
		return
	}

	if err := b.sessionManager.PauseSession(userID); err != nil {
		b.respondWithError(s, i, fmt.Sprintf("Error al pausar el pomodoro: %v", err))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⏸️ Pomodoro Pausado",
		Description: "Tu sesión de pomodoro ha sido pausada. Usa `/pomodoro-resume` para continuar.",
		Color:       0xffaa00,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// handleResumePomodoro maneja el comando de reanudar pomodoro
func (b *Bot) handleResumePomodoro(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID, err := b.getUserID(i)
	if err != nil {
		b.respondWithError(s, i, err.Error())
		return
	}

	if err := b.sessionManager.ResumeSession(userID); err != nil {
		b.respondWithError(s, i, fmt.Sprintf("Error al reanudar el pomodoro: %v", err))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "▶️ Pomodoro Reanudado",
		Description: "Tu sesión de pomodoro ha sido reanudada. ¡Sigue adelante!",
		Color:       0x00ff00,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// handleSkipPomodoro maneja el comando de saltar sesión
func (b *Bot) handleSkipPomodoro(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID, err := b.getUserID(i)
	if err != nil {
		b.respondWithError(s, i, err.Error())
		return
	}

	if err := b.sessionManager.SkipSession(userID); err != nil {
		b.respondWithError(s, i, fmt.Sprintf("Error al saltar la sesión: %v", err))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⏭️ Sesión Saltada",
		Description: "La sesión actual ha sido saltada. Continuando con la siguiente sesión.",
		Color:       0xffaa00,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// handleStatusPomodoro maneja el comando de estado
func (b *Bot) handleStatusPomodoro(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID, err := b.getUserID(i)
	if err != nil {
		b.respondWithError(s, i, err.Error())
		return
	}

	session, err := b.sessionManager.GetSession(userID)
	if err != nil {
		b.respondWithError(s, i, "No tienes una sesión de pomodoro activa. Usa `/pomodoro` para iniciar una.")
		return
	}

	// Obtener información del engine
	engine := session.Engine
	currentSession := engine.GetCurrentSession()
	state := engine.GetState()
	pomodoroCount := engine.GetPomodoroCount()

	// Determinar emoji y título basado en el tipo de sesión
	statusEmoji := "🍅"
	statusTitle := "Sesión de Trabajo"
	statusColor := 0xff6b6b

	// Convertir SessionType a string para comparar
	sessionTypeStr := string(currentSession)

	switch sessionTypeStr {
	case "work":
		statusEmoji = "🍅"
		statusTitle = "Sesión de Trabajo"
		statusColor = 0xff6b6b
	case "short_break":
		statusEmoji = "☕"
		statusTitle = "Descanso Corto"
		statusColor = 0x4ecdc4
	case "long_break":
		statusEmoji = "🏖️"
		statusTitle = "Descanso Largo"
		statusColor = 0x45b7d1
	}

	stateStr := string(state)
	if stateStr == "paused" {
		statusEmoji = "⏸️"
		statusTitle = "Pausado - " + statusTitle
		statusColor = 0xffa726
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s %s", statusEmoji, statusTitle),
		Description: fmt.Sprintf("Sesión de Pomodoro #%d", pomodoroCount+1),
		Color:       statusColor,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Estado", Value: translateState(stateStr), Inline: true},
			{Name: "Sesión Iniciada", Value: session.StartTime.Format("15:04:05"), Inline: true},
			{Name: "Configuración", Value: fmt.Sprintf("Trabajo: %s | Descanso Corto: %s | Descanso Largo: %s",
				config.FormatDuration(session.Config.WorkDuration),
				config.FormatDuration(session.Config.ShortBreak),
				config.FormatDuration(session.Config.LongBreak)), Inline: false},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Usa /pomodoro-stats para estadísticas detalladas",
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// handleStatsPomodoro maneja el comando de estadísticas
func (b *Bot) handleStatsPomodoro(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID, err := b.getUserID(i)
	if err != nil {
		b.respondWithError(s, i, err.Error())
		return
	}

	session, err := b.sessionManager.GetSession(userID)
	if err != nil {
		b.respondWithError(s, i, "No tienes una sesión de pomodoro activa. Usa `/pomodoro` para iniciar una.")
		return
	}

	statsData := session.Engine.GetStats().GetSnapshot()

	// Crear barra de progreso visual para eficiencia
	efficiencyBar := b.createProgressBar(statsData.WorkEfficiency, 20)

	embed := &discordgo.MessageEmbed{
		Title:       "📊 Estadísticas de Pomodoro",
		Description: "Estadísticas de tu sesión actual",
		Color:       0x9b59b6,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🍅 Pomodoros",
				Value:  fmt.Sprintf("Completados: %d\nSaltados: %d", statsData.PomodorosCompleted, statsData.PomodorosSkipped),
				Inline: true,
			},
			{
				Name:   "☕ Descansos",
				Value:  fmt.Sprintf("Completados: %d\nSaltados: %d\nDescansos Largos: %d", statsData.BreaksCompleted, statsData.BreaksSkipped, statsData.LongBreaksCompleted),
				Inline: true,
			},
			{
				Name:   "🔥 Rachas",
				Value:  fmt.Sprintf("Actual: %d\nMejor: %d", statsData.CurrentStreak, statsData.BestStreak),
				Inline: true,
			},
			{
				Name: "⏱️ Tiempo Dedicado",
				Value: fmt.Sprintf("Trabajo: %s\nDescansos: %s\nTotal: %s",
					stats.FormatDuration(statsData.TotalWorkTime),
					stats.FormatDuration(statsData.TotalBreakTime),
					stats.FormatDuration(statsData.SessionDuration)),
				Inline: true,
			},
			{
				Name:   "📈 Eficiencia",
				Value:  fmt.Sprintf("%.1f%%\n[%s]", statsData.WorkEfficiency, efficiencyBar),
				Inline: true,
			},
			{
				Name: "📋 Info de Sesión",
				Value: fmt.Sprintf("Total de Sesiones: %d\nIniciado: %s",
					statsData.TotalSessions,
					session.StartTime.Format("15:04 del 2 Jan")),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "¡Sigue con el excelente trabajo! 🎯",
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// translateState traduce el estado del engine al español
func translateState(state string) string {
	switch state {
	case "running":
		return "ejecutando"
	case "paused":
		return "pausado"
	case "stopped":
		return "detenido"
	case "idle":
		return "inactivo"
	default:
		return state
	}
}

// createProgressBar crea una barra de progreso visual
func (b *Bot) createProgressBar(percentage float64, width int) string {
	if width <= 0 {
		width = 20
	}

	filled := int(percentage / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	var bar string
	for i := 0; i < width; i++ {
		if i < filled {
			if percentage >= 80 {
				bar += "█" // Verde para alta eficiencia
			} else if percentage >= 60 {
				bar += "▓" // Amarillo para eficiencia media
			} else {
				bar += "▒" // Rojo para baja eficiencia
			}
		} else {
			bar += "░"
		}
	}

	return bar
}
