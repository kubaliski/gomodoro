package bot

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// NotificationManager maneja el envío de notificaciones con lógica DM/fallback
type NotificationManager struct {
	session        *discordgo.Session
	dmChannelCache map[string]string // userID -> dmChannelID
	cacheMutex     sync.RWMutex
	welcomeSent    map[string]bool // userID -> sent
	welcomeMutex   sync.RWMutex
}

// NewNotificationManager crea una nueva instancia del notification manager
func NewNotificationManager(session *discordgo.Session) *NotificationManager {
	return &NotificationManager{
		session:        session,
		dmChannelCache: make(map[string]string),
		welcomeSent:    make(map[string]bool),
	}
}

// SendNotification envía notificación a DM con fallback a canal
// Esta es la función principal para todas las notificaciones automáticas
func (nm *NotificationManager) SendNotification(userID, channelID string, embed *discordgo.MessageEmbed, mention string) error {
	// Enviar mensaje de bienvenida si es la primera vez
	nm.sendWelcomeIfNeeded(userID)

	return nm.sendToDM(userID, channelID, embed, mention)
}

// SendToChannel fuerza envío al canal público (para respuestas a comandos)
func (nm *NotificationManager) SendToChannel(channelID string, embed *discordgo.MessageEmbed) error {
	_, err := nm.session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		return fmt.Errorf("failed to send to channel %s: %w", channelID, err)
	}
	log.Printf("📢 Message sent to channel %s", channelID)
	return nil
}

// SendToChannelWithMention envía embed + mention al canal público
func (nm *NotificationManager) SendToChannelWithMention(channelID string, embed *discordgo.MessageEmbed, mention string) error {
	// Enviar embed
	if err := nm.SendToChannel(channelID, embed); err != nil {
		return err
	}

	// Enviar mention si es necesario
	if mention != "" {
		_, err := nm.session.ChannelMessageSend(channelID, mention)
		if err != nil {
			log.Printf("⚠️ Failed to send mention to channel %s: %v", channelID, err)
		}
	}

	return nil
}

// getOrCreateDMChannel obtiene o crea un canal DM para un usuario con cache
func (nm *NotificationManager) getOrCreateDMChannel(userID string) (string, error) {
	// Verificar cache primero
	nm.cacheMutex.RLock()
	if dmChannelID, exists := nm.dmChannelCache[userID]; exists {
		nm.cacheMutex.RUnlock()
		return dmChannelID, nil
	}
	nm.cacheMutex.RUnlock()

	// Crear canal DM
	channel, err := nm.session.UserChannelCreate(userID)
	if err != nil {
		return "", fmt.Errorf("failed to create DM channel for user %s: %w", userID, err)
	}

	// Guardar en cache
	nm.cacheMutex.Lock()
	nm.dmChannelCache[userID] = channel.ID
	nm.cacheMutex.Unlock()

	log.Printf("📱 Created and cached DM channel for user %s", userID)
	return channel.ID, nil
}

// sendToDM intenta enviar a DM, con fallback automático a canal
func (nm *NotificationManager) sendToDM(userID, channelID string, embed *discordgo.MessageEmbed, mention string) error {
	// 1. Intentar obtener/crear canal DM
	dmChannelID, err := nm.getOrCreateDMChannel(userID)
	if err != nil {
		log.Printf("📢 DM unavailable for user %s, using channel fallback: %v", userID, err)
		return nm.sendToChannelFallback(channelID, embed, mention)
	}

	// 2. Intentar enviar embed a DM
	_, err = nm.session.ChannelMessageSendEmbed(dmChannelID, embed)
	if err != nil {
		log.Printf("📢 DM failed for user %s, using channel fallback: %v", userID, err)
		return nm.sendToChannelFallback(channelID, embed, mention)
	}

	// 3. Enviar mention por separado si es necesario
	if mention != "" {
		_, err = nm.session.ChannelMessageSend(dmChannelID, mention)
		if err != nil {
			log.Printf("⚠️ Failed to send DM mention to user %s: %v", userID, err)
			// No hacemos fallback para mention, solo log
		}
	}

	log.Printf("📱 DM notification sent successfully to user %s", userID)
	return nil
}

// sendToChannelFallback envía notificación al canal público como fallback
func (nm *NotificationManager) sendToChannelFallback(channelID string, embed *discordgo.MessageEmbed, mention string) error {
	// Enviar embed
	_, err := nm.session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		return fmt.Errorf("failed to send fallback embed to channel %s: %w", channelID, err)
	}

	// Enviar mention si es necesario
	if mention != "" {
		_, err = nm.session.ChannelMessageSend(channelID, mention)
		if err != nil {
			log.Printf("⚠️ Failed to send fallback mention to channel %s: %v", channelID, err)
		}
	}

	log.Printf("📢 Fallback notification sent to channel %s", channelID)
	return nil
}

// sendWelcomeIfNeeded envía mensaje de bienvenida si es la primera notificación
func (nm *NotificationManager) sendWelcomeIfNeeded(userID string) {
	nm.welcomeMutex.RLock()
	if sent := nm.welcomeSent[userID]; sent {
		nm.welcomeMutex.RUnlock()
		return
	}
	nm.welcomeMutex.RUnlock()

	// Marcar como enviado para evitar duplicados
	nm.welcomeMutex.Lock()
	nm.welcomeSent[userID] = true
	nm.welcomeMutex.Unlock()

	nm.sendWelcomeMessage(userID)
}

// sendWelcomeMessage envía el mensaje de bienvenida inicial
func (nm *NotificationManager) sendWelcomeMessage(userID string) {
	embed := &discordgo.MessageEmbed{
		Title: "🍅 ¡Bienvenido a Pomodoro Bot!",
		Description: "A partir de ahora recibirás todas las notificaciones de pomodoro aquí en tus mensajes privados.\n\n" +
			"Si no puedes recibir mensajes privados, las notificaciones se enviarán automáticamente al canal donde ejecutaste el comando.",
		Color: 0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "💡 Consejo",
				Value:  "Mantén este chat a mano para no perderte las notificaciones de tus sesiones",
				Inline: false,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "¡Que tengas sesiones productivas!",
		},
	}

	dmChannelID, err := nm.getOrCreateDMChannel(userID)
	if err != nil {
		log.Printf("⚠️ Could not send welcome message to user %s: %v", userID, err)
		return
	}

	_, err = nm.session.ChannelMessageSendEmbed(dmChannelID, embed)
	if err != nil {
		log.Printf("⚠️ Failed to send welcome message to user %s: %v", userID, err)
	} else {
		log.Printf("👋 Welcome message sent to user %s", userID)
	}
}

// ClearCache limpia el cache de canales DM (útil para testing o limpieza)
func (nm *NotificationManager) ClearCache() {
	nm.cacheMutex.Lock()
	nm.dmChannelCache = make(map[string]string)
	nm.cacheMutex.Unlock()

	nm.welcomeMutex.Lock()
	nm.welcomeSent = make(map[string]bool)
	nm.welcomeMutex.Unlock()

	log.Printf("🧹 Notification manager cache cleared")
}
