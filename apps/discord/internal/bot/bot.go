package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kubaliski/gomodoro/apps/discord/internal/manager"
)

// Bot representa el bot de Discord con arquitectura modular
type Bot struct {
	session        *discordgo.Session
	sessionManager *manager.SessionManager
	notifier       *NotificationManager
	eventHandler   *EventHandler
	registry       *CommandRegistry
	isRunning      bool
}

// NewBot crea una nueva instancia del bot con todos sus componentes
func NewBot(token string, sessionManager *manager.SessionManager) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Crear componentes modulares
	notifier := NewNotificationManager(session)
	eventHandler := NewEventHandler(sessionManager)
	registry := NewCommandRegistry()

	bot := &Bot{
		session:        session,
		sessionManager: sessionManager,
		notifier:       notifier,
		eventHandler:   eventHandler,
		registry:       registry,
	}

	// Configurar handlers básicos
	bot.setupHandlers()

	return bot, nil
}

// Start inicia el bot y todos sus componentes
func (b *Bot) Start(ctx context.Context) error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord session: %w", err)
	}

	b.isRunning = true

	// Registrar comandos slash
	if err := b.registry.RegisterCommands(b.session); err != nil {
		log.Printf("⚠️ Failed to register slash commands: %v", err)
	}

	// Iniciar rutina de limpieza periódica
	go b.cleanupRoutine(ctx)

	return nil
}

// Stop detiene el bot y limpia todos los recursos
func (b *Bot) Stop() {
	log.Println("🛑 Stopping bot...")
	b.isRunning = false

	// Detener todas las sesiones activas
	activeSessions := b.sessionManager.GetAllActiveSessions()
	for userID := range activeSessions {
		if err := b.sessionManager.StopSession(userID); err != nil {
			log.Printf("⚠️ Error stopping session for user %s: %v", userID, err)
		}
	}

	// Cerrar conexión Discord
	if b.session != nil {
		b.session.Close()
	}

	log.Println("✅ Bot stopped successfully")
}

// setupHandlers configura los event handlers del bot
func (b *Bot) setupHandlers() {
	// Handler para comandos slash
	b.session.AddHandler(b.handleSlashCommand)

	// Handler para cuando el bot se conecta
	b.session.AddHandler(b.handleReady)

	// Registrar event handlers de pomodoro con el session manager
	b.eventHandler.RegisterWithSessionManager(b.notifier)

	log.Printf("🔧 Bot handlers configured successfully")
}

// handleReady maneja el evento cuando el bot está listo
func (b *Bot) handleReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("✅ Bot is ready! Logged in as: %v#%v", r.User.Username, r.User.Discriminator)
	log.Printf("🤖 Bot is connected to %d guilds", len(r.Guilds))
}

// handleSlashCommand maneja todos los comandos slash
func (b *Bot) handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandName := i.ApplicationCommandData().Name
	log.Printf("📝 Received command: /%s from user %s", commandName, b.getInteractionUserID(i))

	switch commandName {
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
	default:
		log.Printf("⚠️ Unknown command: %s", commandName)
		respondWithError(s, i, "Comando no reconocido")
	}
}

// getInteractionUserID obtiene el ID del usuario de la interacción de forma segura
func (b *Bot) getInteractionUserID(i *discordgo.InteractionCreate) string {
	userID, err := getUserID(i)
	if err != nil {
		return "unknown"
	}
	return userID
}

// cleanupRoutine ejecuta limpieza periódica de sesiones inactivas
func (b *Bot) cleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Printf("🧹 Cleanup routine started (every 5 minutes)")

	for {
		select {
		case <-ctx.Done():
			log.Printf("🧹 Cleanup routine stopped")
			return
		case <-ticker.C:
			if b.isRunning {
				b.sessionManager.CleanupInactiveSessions()
			}
		}
	}
}

// GetNotifier expone el notifier para uso en tests o extensiones
func (b *Bot) GetNotifier() *NotificationManager {
	return b.notifier
}

// GetSessionManager expone el session manager
func (b *Bot) GetSessionManager() *manager.SessionManager {
	return b.sessionManager
}

// IsRunning retorna si el bot está ejecutándose
func (b *Bot) IsRunning() bool {
	return b.isRunning
}
