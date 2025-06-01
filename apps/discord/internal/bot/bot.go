package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kubaliski/gomodoro/apps/discord/internal/manager"
	"github.com/kubaliski/pomodoro-core/config"
	"github.com/kubaliski/pomodoro-core/events"
)

// Bot representa el bot de Discord
type Bot struct {
	session        *discordgo.Session
	sessionManager *manager.SessionManager
	isRunning      bool
}

// NewBot crea una nueva instancia del bot
func NewBot(token string, sessionManager *manager.SessionManager) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	bot := &Bot{
		session:        session,
		sessionManager: sessionManager,
	}

	// Configurar handlers
	bot.setupHandlers()

	return bot, nil
}

// Start inicia el bot
func (b *Bot) Start(ctx context.Context) error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord session: %w", err)
	}

	b.isRunning = true

	// Registrar comandos slash
	if err := b.registerSlashCommands(); err != nil {
		log.Printf("Failed to register slash commands: %v", err)
	}

	// Iniciar limpieza periódica de sesiones
	go b.cleanupRoutine(ctx)

	return nil
}

// Stop detiene el bot
func (b *Bot) Stop() {
	b.isRunning = false

	// Detener todas las sesiones activas
	for userID := range b.sessionManager.GetAllActiveSessions() {
		b.sessionManager.StopSession(userID)
	}

	if b.session != nil {
		b.session.Close()
	}
}

// setupHandlers configura los event handlers del bot
func (b *Bot) setupHandlers() {
	// Handler para comandos slash
	b.session.AddHandler(b.handleSlashCommand)

	// Handler para cuando el bot se conecta
	b.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("✅ Bot is ready! Logged in as: %v#%v", r.User.Username, r.User.Discriminator)
	})

	// Registrar handlers de eventos de pomodoro
	b.sessionManager.RegisterEventHandler("pomodoro_completed", b.handlePomodoroCompleted)
	b.sessionManager.RegisterEventHandler("break_completed", b.handleBreakCompleted)
	b.sessionManager.RegisterEventHandler("pomodoro_started", b.handlePomodoroStarted)
	b.sessionManager.RegisterEventHandler("break_started", b.handleBreakStarted)
	b.sessionManager.RegisterEventHandler("timer_reminder", b.handleTimerReminder)
}

// getOrCreateDMChannel obtiene o crea un canal DM para un usuario
func (b *Bot) getOrCreateDMChannel(userID string) (string, error) {
	channel, err := b.session.UserChannelCreate(userID)
	if err != nil {
		return "", fmt.Errorf("failed to create DM channel for user %s: %w", userID, err)
	}
	return channel.ID, nil
}

// sendNotificationWithFallback envía notificación a DM primero, fallback a canal
func (b *Bot) sendNotificationWithFallback(userID, channelID string, embed *discordgo.MessageEmbed, mention string) error {
	// Verificar si hay sesión activa
	_, err := b.sessionManager.GetSession(userID)
	if err != nil {
		// Si no hay sesión activa, usar canal original
		return b.sendToChannel(channelID, embed, mention)
	}

	// Por ahora solo implementamos modo DM con fallback
	// En Fase 2 usaremos session.NotificationMode para diferentes modos
	return b.sendToDM(userID, channelID, embed, mention)
}

// sendToDM intenta enviar a DM, con fallback a canal
func (b *Bot) sendToDM(userID, channelID string, embed *discordgo.MessageEmbed, mention string) error {
	// 1. Intentar obtener/crear canal DM
	dmChannelID, err := b.getOrCreateDMChannel(userID)
	if err != nil {
		log.Printf("⚠️ Failed to create DM channel for user %s: %v. Using fallback.", userID, err)
		return b.sendToChannel(channelID, embed, mention)
	}

	// 2. Actualizar cache de DM en sesión
	if err := b.sessionManager.UpdateSessionDMChannel(userID, dmChannelID); err != nil {
		log.Printf("⚠️ Failed to update DM channel cache: %v", err)
	}

	// 3. Intentar enviar embed a DM
	_, err = b.session.ChannelMessageSendEmbed(dmChannelID, embed)
	if err != nil {
		log.Printf("⚠️ Failed to send DM embed to user %s: %v. Using fallback.", userID, err)
		return b.sendToChannel(channelID, embed, mention)
	}

	// 4. Enviar mention por separado si es necesario
	if mention != "" {
		_, err = b.session.ChannelMessageSend(dmChannelID, mention)
		if err != nil {
			log.Printf("⚠️ Failed to send DM mention to user %s: %v", userID, err)
			// No hacemos fallback para mention, solo log
		}
	}

	log.Printf("✅ DM notification sent successfully to user %s", userID)
	return nil
}

// sendToChannel envía notificación al canal público
func (b *Bot) sendToChannel(channelID string, embed *discordgo.MessageEmbed, mention string) error {
	// Enviar embed
	_, err := b.session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		return fmt.Errorf("failed to send embed to channel %s: %w", channelID, err)
	}

	// Enviar mention si es necesario
	if mention != "" {
		_, err = b.session.ChannelMessageSend(channelID, mention)
		if err != nil {
			log.Printf("⚠️ Failed to send mention to channel %s: %v", channelID, err)
		}
	}

	log.Printf("📢 Channel notification sent successfully to channel %s", channelID)
	return nil
}

// registerSlashCommands registra los comandos slash del bot
func (b *Bot) registerSlashCommands() error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "pomodoro",
			Description: "Iniciar una nueva sesión de pomodoro",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "work",
					Description: "Duración del trabajo en minutos (por defecto: 25)",
					Required:    false,
					MinValue:    func() *float64 { v := 1.0; return &v }(),
					MaxValue:    120,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "short_break",
					Description: "Duración del descanso corto en minutos (por defecto: 5)",
					Required:    false,
					MinValue:    func() *float64 { v := 1.0; return &v }(),
					MaxValue:    30,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "long_break",
					Description: "Duración del descanso largo en minutos (por defecto: 15)",
					Required:    false,
					MinValue:    func() *float64 { v := 5.0; return &v }(),
					MaxValue:    60,
				},
			},
		},
		{
			Name:        "pomodoro-stop",
			Description: "Detener tu sesión de pomodoro actual",
		},
		{
			Name:        "pomodoro-pause",
			Description: "Pausar tu sesión de pomodoro actual",
		},
		{
			Name:        "pomodoro-resume",
			Description: "Reanudar tu sesión de pomodoro pausada",
		},
		{
			Name:        "pomodoro-skip",
			Description: "Saltar el pomodoro o descanso actual",
		},
		{
			Name:        "pomodoro-status",
			Description: "Verificar el estado actual de tu pomodoro",
		},
		{
			Name:        "pomodoro-stats",
			Description: "Ver tus estadísticas de pomodoro",
		},
	}

	for _, cmd := range commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("failed to create command %s: %w", cmd.Name, err)
		}
	}

	return nil
}

// handleSlashCommand maneja los comandos slash
func (b *Bot) handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "pomodoro":
		b.handleStartPomodoro(s, i)
	case "pomodoro-stop":
		b.handleStopPomodoro(s, i)
	case "pomodoro-pause":
		b.handlePausePomodoro(s, i)
	case "pomodoro-resume":
		b.handleResumePomodoro(s, i)
	case "pomodoro-skip":
		b.handleSkipPomodoro(s, i)
	case "pomodoro-status":
		b.handleStatusPomodoro(s, i)
	case "pomodoro-stats":
		b.handleStatsPomodoro(s, i)
	}
}

// getUserID obtiene el ID del usuario de forma segura (funciona en canal y DM)
func (b *Bot) getUserID(i *discordgo.InteractionCreate) (string, error) {
	if i.Member != nil {
		// Comando ejecutado en servidor
		return i.Member.User.ID, nil
	} else if i.User != nil {
		// Comando ejecutado en DM
		return i.User.ID, nil
	}
	return "", fmt.Errorf("no se pudo identificar el usuario")
}

// handleStartPomodoro maneja el comando de iniciar pomodoro
func (b *Bot) handleStartPomodoro(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID, err := b.getUserID(i)
	if err != nil {
		b.respondWithError(s, i, err.Error())
		return
	}

	channelID := i.ChannelID

	// Parsear opciones personalizadas
	cfg := config.DefaultConfig()
	options := i.ApplicationCommandData().Options

	for _, option := range options {
		switch option.Name {
		case "work":
			cfg.WorkDuration = time.Duration(option.IntValue()) * time.Minute
		case "short_break":
			cfg.ShortBreak = time.Duration(option.IntValue()) * time.Minute
		case "long_break":
			cfg.LongBreak = time.Duration(option.IntValue()) * time.Minute
		}
	}

	// Validar configuración
	if err := cfg.Validate(); err != nil {
		b.respondWithError(s, i, fmt.Sprintf("Configuración inválida: %v", err))
		return
	}

	// Iniciar sesión
	session, err := b.sessionManager.StartSession(userID, channelID, cfg)
	if err != nil {
		b.respondWithError(s, i, fmt.Sprintf("Error al iniciar pomodoro: %v", err))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🍅 ¡Pomodoro Iniciado!",
		Description: fmt.Sprintf("Tu sesión de pomodoro ha comenzado con períodos de trabajo de %s.\n\n📱 *Las notificaciones se enviarán a tus mensajes privados*", config.FormatDuration(session.Config.WorkDuration)),
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Duración de Trabajo", Value: config.FormatDuration(session.Config.WorkDuration), Inline: true},
			{Name: "Descanso Corto", Value: config.FormatDuration(session.Config.ShortBreak), Inline: true},
			{Name: "Descanso Largo", Value: config.FormatDuration(session.Config.LongBreak), Inline: true},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Asegúrate de tener los DMs habilitados para recibir notificaciones",
		},
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
	}
}

// handleStopPomodoro maneja el comando de detener pomodoro
func (b *Bot) handleStopPomodoro(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID, err := b.getUserID(i)
	if err != nil {
		b.respondWithError(s, i, err.Error())
		return
	}

	if err := b.sessionManager.StopSession(userID); err != nil {
		b.respondWithError(s, i, fmt.Sprintf("Error al detener el pomodoro: %v", err))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⏹️ Pomodoro Detenido",
		Description: "Tu sesión de pomodoro ha sido detenida.",
		Color:       0xff0000,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// Event handlers para notificaciones de pomodoro (ACTUALIZADOS PARA DM)

func (b *Bot) handlePomodoroCompleted(userID, channelID string, event events.Event) {
	data, ok := event.Data.(events.PomodoroEventData)
	if !ok {
		log.Printf("Invalid event data type for PomodoroCompleted")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🎉 ¡Pomodoro Completado!",
		Description: fmt.Sprintf("¡Excelente trabajo! Has completado el pomodoro #%d", data.Number),
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Duración", Value: config.FormatDuration(data.Duration), Inline: true},
			{Name: "Tiempo Real", Value: config.FormatDuration(data.ActualTime), Inline: true},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	mention := "¡Hora de un descanso! 🧘‍♂️"

	if err := b.sendNotificationWithFallback(userID, channelID, embed, mention); err != nil {
		log.Printf("Error sending pomodoro completed notification: %v", err)
	}
}

func (b *Bot) handleBreakStarted(userID, channelID string, event events.Event) {
	data, ok := event.Data.(events.BreakEventData)
	if !ok {
		log.Printf("Invalid event data type for BreakStarted")
		return
	}

	breakType := "Descanso Corto"
	emoji := "☕"
	if data.IsLongBreak {
		breakType = "Descanso Largo"
		emoji = "🏖️"
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s %s Iniciado", emoji, breakType),
		Description: fmt.Sprintf("Hora de relajarse por %s", config.FormatDuration(data.Duration)),
		Color:       0x0099ff,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	if err := b.sendNotificationWithFallback(userID, channelID, embed, ""); err != nil {
		log.Printf("Error sending break started notification: %v", err)
	}
}

func (b *Bot) handlePomodoroStarted(userID, channelID string, event events.Event) {
	data, ok := event.Data.(events.PomodoroEventData)
	if !ok {
		log.Printf("Invalid event data type for PomodoroStarted")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🍅 ¡Hora de Concentrarse!",
		Description: fmt.Sprintf("Pomodoro #%d iniciado - ¡hora de enfocarse!", data.Number),
		Color:       0xff6b6b,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Duración", Value: config.FormatDuration(data.Duration), Inline: true},
			{Name: "Iniciado", Value: data.StartTime.Format("15:04:05"), Inline: true},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := b.sendNotificationWithFallback(userID, channelID, embed, ""); err != nil {
		log.Printf("Error sending pomodoro started notification: %v", err)
	}
}

func (b *Bot) handleBreakCompleted(userID, channelID string, event events.Event) {
	data, ok := event.Data.(events.BreakEventData)
	if !ok {
		log.Printf("Invalid event data type for BreakCompleted")
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
	}

	mention := "¡De vuelta al trabajo! 💪"

	if err := b.sendNotificationWithFallback(userID, channelID, embed, mention); err != nil {
		log.Printf("Error sending break completed notification: %v", err)
	}
}

func (b *Bot) handleTimerReminder(userID, channelID string, event events.Event) {
	data, ok := event.Data.(events.TimerEventData)
	if !ok {
		return
	}

	remaining := int(data.Remaining.Minutes())

	var message string
	var color int

	switch remaining {
	case 10:
		message = "Quedan 10 minutos"
		color = 0xffaa00
	case 5:
		message = "Quedan 5 minutos"
		color = 0xff6600
	case 1:
		message = "¡Queda 1 minuto!"
		color = 0xff0000
	default:
		return // No reminder needed
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⏰ Recordatorio de Tiempo",
		Description: message,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	if err := b.sendNotificationWithFallback(userID, channelID, embed, ""); err != nil {
		log.Printf("Error sending timer reminder: %v", err)
	}
}

// Helper functions

func (b *Bot) respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	embed := &discordgo.MessageEmbed{
		Title:       "❌ Error",
		Description: message,
		Color:       0xff0000,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

// translateBreakType traduce el tipo de descanso
func translateBreakType(breakType string) string {
	switch breakType {
	case "DESCANSO":
		return "Descanso Corto"
	case "DESCANSO LARGO":
		return "Descanso Largo"
	default:
		return breakType
	}
}

func (b *Bot) cleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.sessionManager.CleanupInactiveSessions()
		}
	}
}
