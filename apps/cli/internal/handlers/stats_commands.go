package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/ui"
)

// StatsCommands maneja todos los comandos relacionados con estadísticas
type StatsCommands struct {
	handler *CLIHandler
}

// NewStatsCommands crea un nuevo handler de comandos de estadísticas
func NewStatsCommands(h *CLIHandler) *StatsCommands {
	return &StatsCommands{handler: h}
}

// ShowDetailedStats muestra estadísticas detalladas con modo interactivo
func (sc *StatsCommands) ShowDetailedStats() {
	if !sc.handler.IsFirstSessionStarted() {
		fmt.Println("❌ Aún no hay estadísticas. Usa 'c' para empezar el primer pomodoro.")
		return
	}

	sc.handler.SetShowingStats(true)
	ui.ClearScreen()

	stats := sc.handler.GetEngine().GetStats()
	config := ui.DefaultStatsConfig()

	// Mostrar estadísticas completas
	statsDisplay := ui.EnhancedStatsDisplay(stats, config)
	fmt.Print(statsDisplay)

	fmt.Println("\n" + ui.Colorize("─────────────────────────────────────────────────────────────", ui.ColorGray, true))
	fmt.Println(ui.Colorize("📋 COMANDOS ADICIONALES:", ui.ColorYellow, true))
	fmt.Println("   • 'compact' - Ver estadísticas compactas")
	fmt.Println("   • 'export' - Exportar datos (próximamente)")
	fmt.Println("   • 'reset' - Reiniciar estadísticas de sesión")
	fmt.Println("   • 'notif-stats' - Estadísticas de notificaciones")
	fmt.Println("   • Enter o 'c' - Volver al timer")
	fmt.Print("Comando stats > ")

	// Loop de comandos de estadísticas
	sc.handleStatsCommands()
}

// ShowCompactStats muestra estadísticas en formato compacto
func (sc *StatsCommands) ShowCompactStats() {
	if !sc.handler.IsFirstSessionStarted() {
		fmt.Println("❌ Aún no hay estadísticas. Usa 'c' para empezar el primer pomodoro.")
		return
	}

	ui.ClearScreen()

	stats := sc.handler.GetEngine().GetStats()
	config := ui.DefaultStatsConfig()
	config.CompactMode = true
	config.ShowGraphs = false
	config.ShowTrends = false

	fmt.Println(ui.Colorize("🍅 ESTADÍSTICAS COMPACTAS", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("─────────────────────────", ui.ColorGray, true))
	fmt.Println()

	compactDisplay := ui.EnhancedStatsDisplay(stats, config)
	fmt.Print(compactDisplay)

	fmt.Println("\n\n📋 'detailed' para ver completas | 'notif-stats' para notificaciones | Enter para volver")
	fmt.Print("Comando stats > ")
}

// ShowQuickStatus muestra un estado rápido del sistema
func (sc *StatsCommands) ShowQuickStatus() {
	if !sc.handler.IsFirstSessionStarted() {
		fmt.Println("📊 Estado: Sistema listo, esperando inicio")
		return
	}

	timerData := sc.handler.GetCurrentTimerData()
	statsData := sc.handler.GetCurrentStatsData()

	fmt.Println()
	fmt.Println(ui.Colorize("📊 ESTADO RÁPIDO", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("──────────────", ui.ColorGray, true))

	// Estado del timer con color contextual
	stateColor := ui.GetTimerStateColor(timerData.State)
	fmt.Printf("⏱️  Timer: %s (%s)\n",
		ui.Colorize(timerData.State, stateColor, true),
		ui.Colorize(timerData.Status, ui.ColorGray, true))

	if timerData.Remaining > 0 {
		remainingTime := time.Duration(timerData.Remaining) * time.Nanosecond
		fmt.Printf("⏰ Restante: %s\n", ui.Colorize(FormatDuration(remainingTime), ui.ColorYellow, true))
	}

	// Progress bar visual
	if timerData.Total > 0 {
		progress := float64(timerData.Total-timerData.Remaining) / float64(timerData.Total)
		progressBar := ui.CreateStyledProgressBar(progress, 20, ui.ClassicProgressBar, true)
		fmt.Printf("📊 Progreso: %s %.1f%%\n", progressBar, progress*100)
	}

	// Stats rápidas con colores
	fmt.Printf("🍅 Completados: %s | 🔥 Racha: %s | ⏱️ Tiempo: %s\n",
		ui.Colorize(fmt.Sprintf("%d", statsData.PomodorosCompleted), ui.ColorGreen, true),
		ui.Colorize(fmt.Sprintf("%d", statsData.CurrentStreak), ui.GetStreakColor(statsData.CurrentStreak), true),
		ui.Colorize(FormatDuration(statsData.TotalWorkTime), ui.ColorBlue, true))

	efficiencyColor := ui.GetEfficiencyColor(statsData.WorkEfficiency)
	fmt.Printf("📈 Eficiencia: %s%.1f%%%s\n",
		string(efficiencyColor), statsData.WorkEfficiency, string(ui.ColorReset))

	// Estado de notificaciones
	notifEnabled := sc.handler.GetNotificationManager().IsEnabled()
	config := sc.handler.GetNotificationManager().GetConfig()
	fmt.Printf("🔔 Notificaciones: %s", sc.enabledStatus(notifEnabled))
	if notifEnabled && config.SoundEnabled {
		fmt.Printf(" (🔊 Vol: %.0f%%)", config.SoundVolume*100)
	}
	fmt.Println()
	fmt.Println()
}

// handleStatsCommands maneja el loop interactivo de comandos de estadísticas
func (sc *StatsCommands) handleStatsCommands() {
	inputChan := sc.handler.GetInputManager().GetInputChannel()

	for {
		select {
		case input := <-inputChan:
			switch strings.TrimSpace(strings.ToLower(input)) {
			case "", "c", "continue", "back", "volver":
				sc.handler.SetShowingStats(false)
				ui.ClearScreen()
				return

			case "compact", "compacto":
				sc.ShowCompactStats()

			case "detailed", "detallado", "full", "completo":
				sc.ShowDetailedStats()
				return

			case "reset", "reiniciar":
				sc.confirmResetStats()

			case "export", "exportar":
				fmt.Println("🚧 Función de exportación próximamente...")
				fmt.Print("Comando stats > ")

			case "notif-stats", "notification-stats":
				sc.handler.GetNotificationCommands().ShowStats()
				fmt.Print("Comando stats > ")

			case "help", "h", "ayuda":
				sc.showStatsHelp()

			default:
				fmt.Printf("❌ Comando '%s' no reconocido en modo stats\n", input)
				fmt.Print("Comando stats > ")
			}
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// confirmResetStats pide confirmación para reiniciar estadísticas
func (sc *StatsCommands) confirmResetStats() {
	fmt.Println()
	fmt.Println(ui.Colorize("⚠️  CONFIRMAR REINICIO DE ESTADÍSTICAS", ui.ColorRed, true))
	fmt.Println("¿Estás seguro de que quieres reiniciar las estadísticas de esta sesión?")
	fmt.Println("Esta acción NO se puede deshacer.")
	fmt.Println()
	fmt.Println("Escribe 'CONFIRMAR' para proceder, o cualquier otra cosa para cancelar:")
	fmt.Print("Confirmación > ")

	inputChan := sc.handler.GetInputManager().GetInputChannel()

	select {
	case input := <-inputChan:
		if strings.TrimSpace(strings.ToUpper(input)) == "CONFIRMAR" {
			fmt.Println(ui.Colorize("✅ Estadísticas reiniciadas", ui.ColorGreen, true))
			fmt.Println("(Nota: Reinicio completo requiere reiniciar la aplicación)")
			// También reiniciar estadísticas de notificaciones
			sc.handler.GetNotificationManager().ResetStats()
			fmt.Println("🔔 Estadísticas de notificaciones también reiniciadas")
		} else {
			fmt.Println(ui.Colorize("❌ Reinicio cancelado", ui.ColorYellow, true))
		}
		time.Sleep(2 * time.Second)
		sc.ShowDetailedStats()
	case <-time.After(30 * time.Second):
		fmt.Println(ui.Colorize("⏰ Tiempo agotado - reinicio cancelado", ui.ColorYellow, true))
		time.Sleep(1 * time.Second)
		sc.ShowDetailedStats()
	}
}

// showStatsHelp muestra ayuda específica para el modo estadísticas
func (sc *StatsCommands) showStatsHelp() {
	fmt.Println()
	fmt.Println(ui.Colorize("📋 AYUDA - MODO ESTADÍSTICAS", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("─────────────────────────────", ui.ColorGray, true))
	fmt.Println()
	fmt.Println("📊 COMANDOS DISPONIBLES:")
	fmt.Println("   • detailed/completo  - Vista detallada con gráficos")
	fmt.Println("   • compact/compacto   - Vista compacta")
	fmt.Println("   • reset/reiniciar    - Reiniciar estadísticas")
	fmt.Println("   • export/exportar    - Exportar datos (próximamente)")
	fmt.Println("   • notif-stats        - Estadísticas de notificaciones")
	fmt.Println("   • help/ayuda         - Esta ayuda")
	fmt.Println("   • c/continue/Enter   - Volver al timer")
	fmt.Println()
	fmt.Println("💡 CONSEJOS:")
	fmt.Println("   • Las estadísticas se actualizan automáticamente")
	fmt.Println("   • Los logros se desbloquean al alcanzar hitos")
	fmt.Println("   • La eficiencia se calcula vs tiempo teórico")
	fmt.Println("   • Las notificaciones también tienen sus propias estadísticas")
	fmt.Println()
	fmt.Print("Comando stats > ")
}

// Helper methods

func (sc *StatsCommands) enabledStatus(enabled bool) string {
	if enabled {
		return ui.Colorize("✅ Habilitado", ui.ColorGreen, true)
	}
	return ui.Colorize("❌ Deshabilitado", ui.ColorRed, true)
}
