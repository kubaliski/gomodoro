package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// CommandRegistry maneja el registro de comandos slash
type CommandRegistry struct {
	commands []*discordgo.ApplicationCommand
}

// NewCommandRegistry crea una nueva instancia del registry
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: []*discordgo.ApplicationCommand{
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
		},
	}
}

// RegisterCommands registra todos los comandos slash con Discord
func (cr *CommandRegistry) RegisterCommands(session *discordgo.Session) error {
	log.Printf("📝 Registering %d slash commands...", len(cr.commands))

	for _, cmd := range cr.commands {
		_, err := session.ApplicationCommandCreate(session.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("failed to create command %s: %w", cmd.Name, err)
		}
		log.Printf("✅ Registered command: /%s", cmd.Name)
	}

	log.Printf("✅ Successfully registered all %d slash commands", len(cr.commands))
	return nil
}

// GetCommands retorna la lista de comandos (útil para testing o limpieza)
func (cr *CommandRegistry) GetCommands() []*discordgo.ApplicationCommand {
	return cr.commands
}
