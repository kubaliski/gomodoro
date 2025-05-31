package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/ui"
)

// UIHelpers maneja todos los elementos de interfaz de usuario y display
type UIHelpers struct {
	handler *CLIHandler
}

// NewUIHelpers crea un nuevo helper de UI
func NewUIHelpers(h *CLIHandler) *UIHelpers {
	return &UIHelpers{handler: h}
}

// ShowConfiguration muestra la configuración inicial del sistema
func (uh *UIHelpers) ShowConfiguration() {
	ui.ClearScreen()
	cfg := uh.handler.GetEngine().GetConfig()

	fmt.Println(ui.Colorize("+================================+", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("|          POMODORO CLI          |", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("+================================+", ui.ColorCyan, true))
	fmt.Println()
	fmt.Println("📋 Configuración:")
	fmt.Printf("   • Trabajo: %s\n", ui.Colorize(ui.FormatDuration(cfg.WorkDuration), ui.ColorRed, true))
	fmt.Printf("   • Descanso corto: %s\n", ui.Colorize(ui.FormatDuration(cfg.ShortBreak), ui.ColorCyan, true))
	fmt.Printf("   • Descanso largo: %s\n", ui.Colorize(ui.FormatDuration(cfg.LongBreak), ui.ColorBlue, true))
	fmt.Printf("   • Descanso largo cada: %s pomodoros\n", ui.Colorize(fmt.Sprintf("%d", cfg.LongBreakInterval), ui.ColorYellow, true))

	// Mostrar estado de notificaciones
	notifConfig := uh.handler.GetNotificationManager().GetConfig()
	fmt.Println()
	fmt.Println("🔔 Notificaciones:")
	fmt.Printf("   • Sonido: %s\n", uh.enabledStatus(notifConfig.SoundEnabled))
	fmt.Printf("   • Volumen: %s%.0f%%%s\n",
		ui.ColorStart(ui.ColorYellow, true), notifConfig.SoundVolume*100, ui.ColorEnd(true))

	fmt.Println()
	fmt.Println("🎮 Controles: (p)ausar (r)eanudar (s)altar (q)salir (h)ayuda")
	fmt.Println("📊 Nuevo: (stats) estadísticas | (compact) vista rápida | (demo) temas")
	fmt.Println("🔊 Audio: (test-sound) probar | (sound-on/off) activar/desactivar")
	fmt.Println("   • Escribe el comando y presiona Enter")
	fmt.Println()
	fmt.Println(ui.Colorize("🚀 Iniciando en 3 segundos...", ui.ColorGreen, true))
	time.Sleep(3 * time.Second)
	ui.ClearScreen()
}

// ShowInitialPrompt muestra el prompt inicial del sistema
func (uh *UIHelpers) ShowInitialPrompt() {
	fmt.Println("✅ Sistema listo. Escribe un comando para empezar:")
	fmt.Println("   • 'c' o Enter para empezar el primer pomodoro")
	fmt.Println("   • 'test-sound' para probar notificaciones")
	fmt.Println("   • 'q' para salir")
	fmt.Println("   • 'h' para ayuda")
	fmt.Print("Comando > ")
}

// DisplayTimerWithStats muestra el timer actual con estadísticas
func (uh *UIHelpers) DisplayTimerWithStats() {
	timerData := uh.handler.GetCurrentTimerData()
	statsData := uh.handler.GetCurrentStatsData()

	// Timer principal con información más clara
	state := timerData.State
	status := timerData.Status

	// Mostrar número de sesión actual para mayor claridad
	sessionInfo := ""
	if state == "TRABAJO" {
		sessionInfo = fmt.Sprintf(" #%d", uh.handler.GetEngine().GetPomodoroCount()+1)
	}

	ui.DisplayTimer(timerData.Remaining, state+sessionInfo, status, timerData.Total)

	// Estadísticas rápidas en la misma línea
	quickStats := fmt.Sprintf("🍅 %d | 🔥 %d | ⏱️ %s",
		statsData.PomodorosCompleted,
		statsData.CurrentStreak,
		FormatDuration(statsData.TotalWorkTime))
	fmt.Printf(" | %s", quickStats)
}

// ShowInlineHelp muestra la ayuda contextual
func (uh *UIHelpers) ShowInlineHelp() {
	fmt.Println()
	fmt.Println(ui.Colorize("🎮 COMANDOS DISPONIBLES", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("─────────────────────", ui.ColorGray, true))

	if uh.handler.IsFirstSessionStarted() {
		fmt.Println("⏱️  CONTROL DEL TIMER:")
		fmt.Println("   • (p)ause    - Pausar timer actual")
		fmt.Println("   • (r)esume   - Reanudar timer pausado")
		fmt.Println("   • (s)kip     - Saltar sesión actual")
		fmt.Println("   • (c)ontinue - Continuar al siguiente")
		fmt.Println()
		fmt.Println("📊 ESTADÍSTICAS:")
		fmt.Println("   • stats      - Ver estadísticas detalladas")
		fmt.Println("   • compact    - Ver estadísticas compactas")
		fmt.Println("   • status     - Estado rápido del timer")
		fmt.Println()
		fmt.Println("🔔 NOTIFICACIONES:")
		fmt.Println("   • test-sound - Probar sonidos")
		fmt.Println("   • notifications - Ver configuración")
		fmt.Println("   • sound-on/off - Activar/desactivar sonido")
		fmt.Println("   • notif-stats - Estadísticas de notificaciones")
		fmt.Println("   • vol+/vol- - Ajustar volumen")
		fmt.Println()
		fmt.Println("🎨 EXTRAS:")
		fmt.Println("   • demo       - Demostración de temas")
		fmt.Println("   • test       - Prueba de características")
		fmt.Println("   • (h)elp     - Esta ayuda")
		fmt.Println("   • (q)uit     - Salir del programa")
	} else {
		fmt.Println("🚀 INICIO:")
		fmt.Println("   • (c)ontinue - Empezar el primer pomodoro")
		fmt.Println("   • (h)elp     - Mostrar esta ayuda")
		fmt.Println("   • (q)uit     - Salir del programa")
		fmt.Println()
		fmt.Println("🔊 AUDIO:")
		fmt.Println("   • test-sound - Probar sistema de sonidos")
		fmt.Println("   • notifications - Ver configuración de audio")
		fmt.Println()
		fmt.Println("🎨 PREVIEW:")
		fmt.Println("   • demo       - Ver demostración de temas")
		fmt.Println("   • test       - Probar características")
		fmt.Println()
		fmt.Println("💡 Después de empezar tendrás más comandos disponibles")
	}
	fmt.Println()
}

// ShowThemeDemo muestra una demostración de temas disponibles
func (uh *UIHelpers) ShowThemeDemo() {
	ui.ClearScreen()

	fmt.Println(ui.Colorize("🎨 DEMOSTRACIÓN DE TEMAS", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("═════════════════════════", ui.ColorGray, true))
	fmt.Println()

	themes := ui.GetAvailableThemes()

	for i, theme := range themes {
		fmt.Printf("%d. %s\n", i+1, ui.Colorize(theme.Name, theme.Primary, true))
		fmt.Printf("   🍅 Trabajo: %s\n", theme.ApplyTheme("25:00", "primary", true))
		fmt.Printf("   🧘 Descanso: %s\n", theme.ApplyTheme("05:00", "info", true))
		fmt.Printf("   ✅ Éxito: %s\n", theme.ApplyTheme("Completado", "success", true))
		fmt.Printf("   ⚠️  Advertencia: %s\n", theme.ApplyTheme("Pausado", "warning", true))
		fmt.Printf("   ❌ Error: %s\n", theme.ApplyTheme("Error", "error", true))

		// Demostrar barra de progreso
		progressDemo := ui.CreateStyledProgressBar(0.7, 20, ui.ClassicProgressBar, true)
		fmt.Printf("   📊 Progreso: %s 70%%\n", progressDemo)
		fmt.Println()
	}

	fmt.Println("💡 Nota: El tema actual es 'Clásico'. Personalización de temas próximamente.")
	fmt.Println("🔊 Consejo: Usa 'test-sound' para probar notificaciones de audio.")
	fmt.Println("\nPresiona Enter para continuar...")

	uh.waitForInput(30)
}

// RunFeatureTest ejecuta una prueba completa de características
func (uh *UIHelpers) RunFeatureTest() {
	ui.ClearScreen()

	fmt.Println(ui.Colorize("🧪 PRUEBA DE CARACTERÍSTICAS", ui.ColorYellow, true))
	fmt.Println(ui.Colorize("═══════════════════════════", ui.ColorGray, true))
	fmt.Println()

	// Test de colores
	fmt.Println("🎨 Test de colores:")
	colors := []struct {
		name  string
		color ui.Color
	}{
		{"Rojo", ui.ColorRed},
		{"Verde", ui.ColorGreen},
		{"Azul", ui.ColorBlue},
		{"Amarillo", ui.ColorYellow},
		{"Cian", ui.ColorCyan},
		{"Magenta", ui.ColorMagenta},
		{"Naranja", ui.ColorOrange},
		{"Púrpura", ui.ColorPurple},
	}

	for _, c := range colors {
		fmt.Printf("   %s ", ui.Colorize("●", c.color, true))
	}
	fmt.Println()
	fmt.Println()

	// Test de barras de progreso
	fmt.Println("📊 Test de barras de progreso:")
	progresses := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	styles := []ui.ProgressBarStyle{
		ui.ClassicProgressBar,
		ui.MinimalProgressBar,
		ui.RetroProgressBar,
	}

	styleNames := []string{"Clásico", "Minimal", "Retro"}

	for i, style := range styles {
		fmt.Printf("   %s:\n", styleNames[i])
		for _, prog := range progresses {
			bar := ui.CreateStyledProgressBar(prog, 30, style, true)
			fmt.Printf("      %s %.0f%%\n", bar, prog*100)
		}
		fmt.Println()
	}

	// Test de logros simulados
	fmt.Println("🏆 Test de logros:")
	achievements := []struct{ icon, desc string }{
		{"🌱", "Primer paso - Sesión iniciada"},
		{"🔥", "En racha - 3 pomodoros consecutivos"},
		{"⚡", "Velocista - 90%+ eficiencia"},
		{"🎯", "Enfocado - Sin saltar descansos"},
		{"🏅", "Constante - 10 pomodoros en una sesión"},
	}

	for _, achievement := range achievements {
		fmt.Printf("   %s %s\n", achievement.icon,
			ui.Colorize(achievement.desc, ui.ColorBrightGreen, true))
	}
	fmt.Println()

	// Test de eficiencia por colores
	fmt.Println("📈 Test de eficiencia por colores:")
	efficiencies := []int{95, 80, 65, 45, 25}
	for _, eff := range efficiencies {
		color := ui.GetEfficiencyColor(float64(eff))
		bar := ui.CreateStyledProgressBar(float64(eff)/100.0, 20, ui.ProgressBarStyle{
			FilledChar: "█", EmptyChar: "░",
			FilledColor: color, EmptyColor: ui.ColorGray,
			BorderColor: ui.ColorWhite,
		}, true)
		fmt.Printf("   %s %s%d%%%s\n", bar, string(color), eff, string(ui.ColorReset))
	}
	fmt.Println()

	// Test de rachas
	fmt.Println("🔥 Test de colores de racha:")
	streaks := []int{0, 1, 3, 5, 10, 15}
	for _, streak := range streaks {
		color := ui.GetStreakColor(streak)
		fmt.Printf("   Racha %s%2d%s: %s\n",
			string(color), streak, string(ui.ColorReset),
			ui.Colorize(strings.Repeat("🔥", min(streak, 5)), color, true))
	}
	fmt.Println()

	// Test de notificaciones
	fmt.Println("🔔 Test de notificadores disponibles:")
	registeredTypes := uh.handler.GetNotificationManager().GetRegisteredNotifiers()
	for _, notifType := range registeredTypes {
		fmt.Printf("   %s %s\n", "✅", notifType)
	}
	if len(registeredTypes) == 0 {
		fmt.Println("   ❌ No hay notificadores registrados")
	}
	fmt.Println()

	fmt.Println("✅ Prueba completada.")
	fmt.Println("💡 Usa 'test-sound' para probar específicamente las notificaciones.")
	fmt.Println("Presiona Enter para continuar...")

	uh.waitForInput(30)
}

// Helper methods

// enabledStatus retorna el estado formateado de habilitado/deshabilitado
func (uh *UIHelpers) enabledStatus(enabled bool) string {
	if enabled {
		return ui.Colorize("✅ Habilitado", ui.ColorGreen, true)
	}
	return ui.Colorize("❌ Deshabilitado", ui.ColorRed, true)
}

// waitForInput espera input del usuario con timeout
func (uh *UIHelpers) waitForInput(timeoutSeconds int) {
	inputChan := uh.handler.GetInputManager().GetInputChannel()

	select {
	case <-inputChan:
		return
	case <-time.After(time.Duration(timeoutSeconds) * time.Second):
		return
	}
}
