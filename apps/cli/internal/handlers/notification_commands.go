package handlers

import (
	"fmt"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/notifications"
	"github.com/kubaliski/pomodoro-cli/internal/ui"
)

// NotificationCommands maneja todos los comandos relacionados con notificaciones
type NotificationCommands struct {
	handler *CLIHandler
}

// NewNotificationCommands crea un nuevo handler de comandos de notificaciones
func NewNotificationCommands(h *CLIHandler) *NotificationCommands {
	return &NotificationCommands{handler: h}
}

// TestNotifications ejecuta una prueba completa de sonidos
func (nc *NotificationCommands) TestNotifications() {
	fmt.Println("🧪 Probando notificaciones de sonido...")
	fmt.Println("   (Nota: Escucharás diferentes tipos de sonidos)")
	fmt.Println()

	// Guardar configuración original
	originalConfig := nc.handler.GetNotificationManager().GetConfig()

	// Crear configuración temporal para el test (habilitar todos los eventos)
	testConfig := originalConfig.Clone()
	testConfig.SystemNotifications = true
	testConfig.PomodoroNotifications = true
	testConfig.BreakNotifications = true
	testConfig.EarlyAlerts = true
	testConfig.UrgentAlerts = true
	testConfig.SoundEnabled = true // Asegurar que el sonido esté habilitado

	// Aplicar configuración temporal
	if err := nc.handler.GetNotificationManager().UpdateConfig(testConfig); err != nil {
		fmt.Printf("❌ Error configurando test: %v\n", err)
		return
	}

	// Probar diferentes tipos de sonido con los EventType correctos
	sounds := []struct {
		name  string
		event notifications.EventType
		delay int // segundos entre tests
	}{
		{"Inicio de sesión", notifications.EventSessionStarted, 1},
		{"Alerta temprana (5 min)", notifications.EventEarlyAlert, 1},
		{"Alerta urgente (1 min)", notifications.EventUrgentAlert, 2},
		{"Pomodoro completado", notifications.EventPomodoroCompleted, 2},
		{"Descanso completado", notifications.EventBreakCompleted, 1},
		{"Timer pausado", notifications.EventTimerPaused, 1},
		{"Timer reanudado", notifications.EventTimerResumed, 1},
	}

	for i, sound := range sounds {
		fmt.Printf("   %d. Probando: %s... ", i+1, sound.name)

		// Usar NotificationRequest directamente para más control
		request := notifications.NotificationRequest{
			Event:    sound.event,
			Title:    "🧪 Test",
			Message:  fmt.Sprintf("Probando %s", sound.name),
			Priority: notifications.PriorityNormal,
			Types:    []notifications.NotificationType{notifications.TypeSound}, // Solo sonido para test
		}

		responses := nc.handler.GetNotificationManager().Notify(request)

		success := false
		for _, resp := range responses {
			if resp.Success {
				success = true
				break
			}
		}

		if success {
			fmt.Println("✅")
		} else {
			fmt.Println("❌")
		}

		time.Sleep(time.Duration(sound.delay) * time.Second) // Pausa entre tests
	}

	// Restaurar configuración original
	if err := nc.handler.GetNotificationManager().UpdateConfig(originalConfig); err != nil {
		fmt.Printf("⚠️ Warning: No se pudo restaurar configuración original: %v\n", err)
	}

	fmt.Println()
	fmt.Println("🎉 Test completado!")
	fmt.Println("💡 Usa 'sound-on/off' para activar/desactivar")
	fmt.Println("💡 Usa 'vol+/vol-' para ajustar volumen")
}

// ShowSettings muestra la configuración actual de notificaciones
func (nc *NotificationCommands) ShowSettings() {
	config := nc.handler.GetNotificationManager().GetConfig()
	registeredTypes := nc.handler.GetNotificationManager().GetRegisteredNotifiers()

	fmt.Println()
	fmt.Println(ui.Colorize("🔔 CONFIGURACIÓN DE NOTIFICACIONES", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("─────────────────────────────────", ui.ColorGray, true))
	fmt.Println()

	// Estado general
	fmt.Printf("🔊 Estado: %s\n", nc.enabledStatus(nc.handler.GetNotificationManager().IsEnabled()))
	fmt.Println()

	// Tipos disponibles
	fmt.Println("📋 Tipos disponibles:")
	for _, notifType := range registeredTypes {
		switch notifType {
		case notifications.TypeSound:
			fmt.Printf("   • Sonido: %s (Vol: %.0f%%)\n",
				nc.enabledStatus(config.SoundEnabled), config.SoundVolume*100)
		case notifications.TypeSystem:
			fmt.Printf("   • Sistema: %s\n", nc.enabledStatus(config.SystemEnabled))
		case notifications.TypeVisual:
			fmt.Printf("   • Visual: %s\n", nc.enabledStatus(config.VisualEnabled))
		}
	}
	fmt.Println()

	// Eventos configurados
	fmt.Println("📅 Eventos:")
	fmt.Printf("   • Pomodoros completados: %s\n", nc.enabledStatus(config.PomodoroNotifications))
	fmt.Printf("   • Descansos completados: %s\n", nc.enabledStatus(config.BreakNotifications))
	fmt.Printf("   • Alertas tempranas (5 min): %s\n", nc.enabledStatus(config.EarlyAlerts))
	fmt.Printf("   • Alertas urgentes (1 min): %s\n", nc.enabledStatus(config.UrgentAlerts))
	fmt.Printf("   • Eventos del sistema: %s\n", nc.enabledStatus(config.SystemNotifications))
	fmt.Println()

	// Horarios silenciosos
	if config.QuietHours.Enabled {
		fmt.Printf("🌙 Horario silencioso: %s - %s\n",
			config.QuietHours.StartTime, config.QuietHours.EndTime)
		if config.IsInQuietHours() {
			fmt.Println("   • Estado: 🌙 Actualmente en horario silencioso")
		} else {
			fmt.Println("   • Estado: ☀️ Horario normal")
		}
		fmt.Println()
	}

	// Comandos disponibles
	fmt.Println(ui.Colorize("🎮 Comandos disponibles:", ui.ColorYellow, true))
	fmt.Println("   • sound-on/off - Activar/desactivar sonido")
	fmt.Println("   • vol+/vol- - Ajustar volumen")
	fmt.Println("   • test-sound - Probar sonidos")
	fmt.Println("   • notif-stats - Ver estadísticas")
	fmt.Println("   • system-on/off - Activar/desactivar notificaciones del sistema")
	fmt.Println("   • alerts-on/off - Activar/desactivar todas las alertas")
	fmt.Println()
}

// ToggleSound activa o desactiva el sonido
func (nc *NotificationCommands) ToggleSound(enabled bool) {
	config := nc.handler.GetNotificationManager().GetConfig()
	config.SoundEnabled = enabled

	if err := nc.handler.GetNotificationManager().UpdateConfig(config); err != nil {
		fmt.Printf("❌ Error actualizando configuración: %v\n", err)
		return
	}

	status := "🔇 deshabilitado"
	if enabled {
		status = "🔊 habilitado"
		// Reproducir sonido de confirmación
		nc.handler.GetNotificationManager().QuickNotify(
			notifications.EventCustomAlert,
			"🔊 Sonido habilitado",
			"Las notificaciones de sonido están activas",
			notifications.PriorityNormal,
		)
	}

	fmt.Printf("🔊 Sonido %s\n", status)
}

// ToggleSystemNotifications activa o desactiva las notificaciones del sistema
func (nc *NotificationCommands) ToggleSystemNotifications(enabled bool) {
	config := nc.handler.GetNotificationManager().GetConfig()
	config.SystemNotifications = enabled

	if err := nc.handler.GetNotificationManager().UpdateConfig(config); err != nil {
		fmt.Printf("❌ Error actualizando configuración: %v\n", err)
		return
	}

	status := "❌ deshabilitadas"
	if enabled {
		status = "✅ habilitadas"
	}

	fmt.Printf("🔔 Notificaciones del sistema %s\n", status)
	if enabled {
		fmt.Println("💡 Ahora recibirás notificaciones para pause/resume/start")
	}
}

// ToggleAllAlerts activa o desactiva todas las alertas
func (nc *NotificationCommands) ToggleAllAlerts(enabled bool) {
	config := nc.handler.GetNotificationManager().GetConfig()
	config.EarlyAlerts = enabled
	config.UrgentAlerts = enabled

	if err := nc.handler.GetNotificationManager().UpdateConfig(config); err != nil {
		fmt.Printf("❌ Error actualizando configuración: %v\n", err)
		return
	}

	status := "❌ deshabilitadas"
	if enabled {
		status = "✅ habilitadas"
	}

	fmt.Printf("⚠️ Alertas de tiempo %s\n", status)
	if enabled {
		fmt.Println("💡 Recibirás alertas a los 5 min, 2 min, 1 min y 30 seg")
	}
}

// AdjustVolume ajusta el volumen de las notificaciones
func (nc *NotificationCommands) AdjustVolume(delta float64) {
	config := nc.handler.GetNotificationManager().GetConfig()
	newVolume := config.SoundVolume + delta

	// Limitar entre 0.0 y 1.0
	if newVolume < 0.0 {
		newVolume = 0.0
	} else if newVolume > 1.0 {
		newVolume = 1.0
	}

	config.SoundVolume = newVolume

	if err := nc.handler.GetNotificationManager().UpdateConfig(config); err != nil {
		fmt.Printf("❌ Error actualizando volumen: %v\n", err)
		return
	}

	fmt.Printf("🔊 Volumen: %.0f%%", newVolume*100)

	// Reproducir sonido de prueba con el nuevo volumen
	if config.SoundEnabled {
		fmt.Print(" - Probando...")
		nc.handler.GetNotificationManager().QuickNotify(
			notifications.EventCustomAlert,
			"🔊 Test de volumen",
			fmt.Sprintf("Volumen ajustado a %.0f%%", newVolume*100),
			notifications.PriorityNormal,
		)
		fmt.Println(" ✅")
	} else {
		fmt.Println(" (sonido deshabilitado)")
	}
}

// ShowStats muestra las estadísticas de notificaciones
func (nc *NotificationCommands) ShowStats() {
	fmt.Println()
	fmt.Println(ui.Colorize("📊 ESTADÍSTICAS DE NOTIFICACIONES", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("─────────────────────────────────", ui.ColorGray, true))
	nc.handler.GetNotificationManager().PrintStats()
	fmt.Println()
}

// TestSpecificSound prueba un tipo específico de sonido
func (nc *NotificationCommands) TestSpecificSound(soundType string) {
	fmt.Printf("🎵 Probando sonido: %s... ", soundType)

	var event notifications.EventType
	switch soundType {
	case "success":
		event = notifications.EventPomodoroCompleted
	case "gentle":
		event = notifications.EventBreakCompleted
	case "warning":
		event = notifications.EventEarlyAlert
	case "urgent":
		event = notifications.EventUrgentAlert
	case "start":
		event = notifications.EventSessionStarted
	case "pause":
		event = notifications.EventTimerPaused
	case "resume":
		event = notifications.EventTimerResumed
	default:
		event = notifications.EventCustomAlert
	}

	// Guardar configuración original y habilitar temporalmente
	originalConfig := nc.handler.GetNotificationManager().GetConfig()
	testConfig := originalConfig.Clone()
	testConfig.SystemNotifications = true
	testConfig.SoundEnabled = true

	if err := nc.handler.GetNotificationManager().UpdateConfig(testConfig); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	response := nc.handler.GetNotificationManager().QuickNotify(
		event,
		"🎵 Test Específico",
		fmt.Sprintf("Probando sonido %s", soundType),
		notifications.PriorityNormal,
	)

	// Restaurar configuración
	nc.handler.GetNotificationManager().UpdateConfig(originalConfig)

	success := false
	for _, resp := range response {
		if resp.Success {
			success = true
			break
		}
	}

	if success {
		fmt.Println("✅")
	} else {
		fmt.Println("❌")
	}
}

// ShowAvailableProfiles muestra los perfiles de configuración disponibles
func (nc *NotificationCommands) ShowAvailableProfiles() {
	fmt.Println()
	fmt.Println(ui.Colorize("👤 PERFILES DE CONFIGURACIÓN", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("─────────────────────────────", ui.ColorGray, true))

	profiles := notifications.GetPredefinedProfiles()
	currentConfig := nc.handler.GetNotificationManager().GetConfig()

	for i, profile := range profiles {
		fmt.Printf("%d. %s\n", i+1, ui.Colorize(profile.Name, ui.ColorYellow, true))
		fmt.Printf("   %s\n", profile.Description)

		// Mostrar configuración resumida
		fmt.Printf("   🔊 Sonido: %s | 🔔 Sistema: %s | ⚠️ Alertas: %s\n",
			nc.enabledStatusShort(profile.Config.SoundEnabled),
			nc.enabledStatusShort(profile.Config.SystemNotifications),
			nc.enabledStatusShort(profile.Config.EarlyAlerts && profile.Config.UrgentAlerts))

		// Indicar si es el perfil activo (aproximación)
		if nc.isConfigSimilar(currentConfig, &profile.Config) {
			fmt.Printf("   %s\n", ui.Colorize("← ACTIVO", ui.ColorGreen, true))
		}
		fmt.Println()
	}

	fmt.Println("💡 Usa 'profile <nombre>' para cambiar perfil")
	fmt.Println("💡 Perfiles disponibles: work, home, focus, silent")
}

// ApplyProfile aplica un perfil de configuración
func (nc *NotificationCommands) ApplyProfile(profileName string) {
	profiles := notifications.GetPredefinedProfiles()

	var selectedProfile *notifications.Profile
	for _, profile := range profiles {
		if profile.Name == profileName {
			selectedProfile = &profile
			break
		}
	}

	if selectedProfile == nil {
		fmt.Printf("❌ Perfil '%s' no encontrado\n", profileName)
		fmt.Println("💡 Perfiles disponibles: work, home, focus, silent")
		return
	}

	if err := nc.handler.GetNotificationManager().UpdateConfig(&selectedProfile.Config); err != nil {
		fmt.Printf("❌ Error aplicando perfil: %v\n", err)
		return
	}

	fmt.Printf("✅ Perfil '%s' aplicado\n", ui.Colorize(selectedProfile.Name, ui.ColorGreen, true))
	fmt.Printf("📝 %s\n", selectedProfile.Description)

	// Mostrar configuración aplicada
	fmt.Println("\n🔧 Configuración actual:")
	fmt.Printf("   🔊 Sonido: %s\n", nc.enabledStatus(selectedProfile.Config.SoundEnabled))
	fmt.Printf("   🔔 Sistema: %s\n", nc.enabledStatus(selectedProfile.Config.SystemNotifications))
	fmt.Printf("   ⚠️ Alertas: %s\n", nc.enabledStatus(selectedProfile.Config.EarlyAlerts))
}

// Helper methods

func (nc *NotificationCommands) enabledStatus(enabled bool) string {
	if enabled {
		return ui.Colorize("✅ Habilitado", ui.ColorGreen, true)
	}
	return ui.Colorize("❌ Deshabilitado", ui.ColorRed, true)
}

func (nc *NotificationCommands) enabledStatusShort(enabled bool) string {
	if enabled {
		return ui.Colorize("✅", ui.ColorGreen, true)
	}
	return ui.Colorize("❌", ui.ColorRed, true)
}

// isConfigSimilar compara configuraciones para determinar si son similares
func (nc *NotificationCommands) isConfigSimilar(config1, config2 *notifications.Config) bool {
	return config1.SoundEnabled == config2.SoundEnabled &&
		config1.SystemNotifications == config2.SystemNotifications &&
		config1.EarlyAlerts == config2.EarlyAlerts &&
		config1.UrgentAlerts == config2.UrgentAlerts
}
