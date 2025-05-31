package handlers

import (
	"fmt"
	"os"

	"github.com/kubaliski/pomodoro-core/engine"
)

// CommandProcessor maneja el procesamiento y routing de comandos
type CommandProcessor struct {
	handler *CLIHandler
}

// NewCommandProcessor crea un nuevo procesador de comandos
func NewCommandProcessor(h *CLIHandler) *CommandProcessor {
	return &CommandProcessor{handler: h}
}

// ProcessCommand procesa un comando de entrada y lo rutea al handler apropiado
func (cp *CommandProcessor) ProcessCommand(input string) {
	// Mostrar el comando escrito
	fmt.Printf("%s\n", input)

	switch input {
	// Control básico del timer
	case "p", "pause":
		cp.handlePause()
	case "r", "resume":
		cp.handleResume()
	case "s", "skip":
		cp.handleSkip()
	case "q", "quit":
		cp.handleQuit()
	case "h", "help":
		cp.handler.GetUIHelpers().ShowInlineHelp()

	// Estadísticas
	case "stats", "estadisticas":
		cp.handler.GetStatsCommands().ShowDetailedStats()
	case "compact", "compacto":
		cp.handler.GetStatsCommands().ShowCompactStats()
	case "status", "estado":
		cp.handler.GetStatsCommands().ShowQuickStatus()

	// Notificaciones
	case "test-sound", "test-audio":
		cp.handler.GetNotificationCommands().TestNotifications()
	case "notifications", "notif", "notificaciones":
		cp.handler.GetNotificationCommands().ShowSettings()
	case "sound-on", "audio-on":
		cp.handler.GetNotificationCommands().ToggleSound(true)
	case "sound-off", "audio-off":
		cp.handler.GetNotificationCommands().ToggleSound(false)
	case "notif-stats", "notification-stats":
		cp.handler.GetNotificationCommands().ShowStats()
	case "volume-up", "vol+":
		cp.handler.GetNotificationCommands().AdjustVolume(0.1)
	case "volume-down", "vol-":
		cp.handler.GetNotificationCommands().AdjustVolume(-0.1)

	// UI y demos
	case "demo", "themes", "temas":
		cp.handler.GetUIHelpers().ShowThemeDemo()
	case "test", "prueba":
		cp.handler.GetUIHelpers().RunFeatureTest()

	// Continuar
	case "c", "continue", "":
		cp.handleContinue()

	default:
		cp.handleUnknownCommand(input)
	}

	cp.showPromptIfNeeded()
}

// Timer control commands

func (cp *CommandProcessor) handlePause() {
	if cp.handler.IsFirstSessionStarted() {
		if err := cp.handler.GetEngine().Pause(); err != nil {
			fmt.Printf("❌ Error pausando: %v\n", err)
		}
	} else {
		fmt.Println("❌ Aún no hay sesión iniciada. Usa 'c' para empezar.")
	}
}

func (cp *CommandProcessor) handleResume() {
	if cp.handler.IsFirstSessionStarted() {
		if err := cp.handler.GetEngine().Resume(); err != nil {
			fmt.Printf("❌ Error reanudando: %v\n", err)
		}
	} else {
		fmt.Println("❌ Aún no hay sesión iniciada. Usa 'c' para empezar.")
	}
}

func (cp *CommandProcessor) handleSkip() {
	if cp.handler.IsFirstSessionStarted() {
		if err := cp.handler.GetEngine().Skip(); err != nil {
			fmt.Printf("❌ Error saltando: %v\n", err)
		}
	} else {
		fmt.Println("❌ Aún no hay sesión iniciada. Usa 'c' para empezar.")
	}
}

func (cp *CommandProcessor) handleQuit() {
	fmt.Println("👋 Saliendo...")
	cp.handler.GetEngine().Stop()
	os.Exit(0)
}

func (cp *CommandProcessor) handleContinue() {
	// Si es la primera vez, iniciar primera sesión
	if !cp.handler.IsFirstSessionStarted() && cp.handler.GetEngine().GetState() == engine.StateIdle {
		if err := cp.handler.GetEngine().StartFirstSession(); err != nil {
			fmt.Printf("❌ Error iniciando sesión: %v\n", err)
		}
	}
	// Si ya hay sesión corriendo, no hacer nada (el engine maneja las transiciones)
}

func (cp *CommandProcessor) handleUnknownCommand(input string) {
	fmt.Printf("❌ Comando '%s' no reconocido.\n", input)
	fmt.Println("💡 Usa 'h' para ver comandos disponibles")
}

func (cp *CommandProcessor) showPromptIfNeeded() {
	// Nuevo prompt para el siguiente comando (solo si no estamos en sesión activa)
	if !cp.handler.IsFirstSessionStarted() || cp.handler.GetEngine().GetState() == engine.StateIdle {
		fmt.Print("Comando > ")
	}
}
