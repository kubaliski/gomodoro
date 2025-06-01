package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// getUserID obtiene el ID del usuario de forma segura (funciona en canal y DM)
func getUserID(i *discordgo.InteractionCreate) (string, error) {
	if i.Member != nil {
		// Comando ejecutado en servidor
		return i.Member.User.ID, nil
	} else if i.User != nil {
		// Comando ejecutado en DM
		return i.User.ID, nil
	}
	return "", fmt.Errorf("no se pudo identificar el usuario")
}

// respondWithError envía una respuesta de error ephemeral
func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
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

// createProgressBar crea una barra de progreso visual
func createProgressBar(percentage float64, width int) string {
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
