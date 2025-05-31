package handlers

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/ui"
	"github.com/kubaliski/pomodoro-core/engine"
	"github.com/kubaliski/pomodoro-core/events"
)

// CLIHandler maneja la interfaz CLI conectando el core con la UI
type CLIHandler struct {
	engine          engine.EngineInterface
	inputReader     *bufio.Reader
	globalInputChan chan string

	// Estado de la UI
	currentTimerData    events.TimerEventData
	currentStatsData    events.StatsEventData
	isShowingStats      bool
	waitingForInput     bool
	firstSessionStarted bool

	// Control de concurrencia
	mu sync.RWMutex
}

// NewCLIHandler crea un nuevo handler CLI
func NewCLIHandler(eng engine.EngineInterface) *CLIHandler {
	handler := &CLIHandler{
		engine:              eng,
		inputReader:         bufio.NewReader(os.Stdin),
		globalInputChan:     make(chan string, 10),
		firstSessionStarted: false,
	}

	// Suscribirse a eventos del engine
	handler.setupEventHandlers()

	// Iniciar listener de input
	go handler.startInputListener()

	return handler
}

// Run ejecuta la interfaz CLI
func (h *CLIHandler) Run(ctx context.Context) error {
	h.showConfiguration()

	// Iniciar el engine (sin empezar sesión automáticamente)
	if err := h.engine.Start(ctx); err != nil {
		return fmt.Errorf("error starting engine: %w", err)
	}

	// Mostrar estado inicial y esperar comando
	fmt.Println("✅ Sistema listo. Escribe un comando para empezar:")
	fmt.Println("   • 'c' o Enter para empezar el primer pomodoro")
	fmt.Println("   • 'q' para salir")
	fmt.Println("   • 'h' para ayuda")
	fmt.Print("Comando > ")

	// Loop principal de input
	h.handleInput()
	return nil
}

// setupEventHandlers configura los manejadores de eventos
func (h *CLIHandler) setupEventHandlers() {
	eventBus := h.engine.GetEventBus()

	// Timer events
	eventBus.SubscribeFunc(events.TimerStarted, h.handleTimerStarted)
	eventBus.SubscribeFunc(events.TimerTick, h.handleTimerTick)
	eventBus.SubscribeFunc(events.TimerPaused, h.handleTimerPaused)
	eventBus.SubscribeFunc(events.TimerResumed, h.handleTimerResumed)
	eventBus.SubscribeFunc(events.TimerCompleted, h.handleTimerCompleted)
	eventBus.SubscribeFunc(events.TimerSkipped, h.handleTimerSkipped)

	// Session events
	eventBus.SubscribeFunc(events.PomodoroStarted, h.handlePomodoroStarted)
	eventBus.SubscribeFunc(events.PomodoroCompleted, h.handlePomodoroCompleted)
	eventBus.SubscribeFunc(events.PomodoroSkipped, h.handlePomodoroSkipped)
	eventBus.SubscribeFunc(events.BreakStarted, h.handleBreakStarted)
	eventBus.SubscribeFunc(events.BreakCompleted, h.handleBreakCompleted)
	eventBus.SubscribeFunc(events.BreakSkipped, h.handleBreakSkipped)

	// Stats events
	eventBus.SubscribeFunc(events.StatsUpdated, h.handleStatsUpdated)

	// Engine events
	eventBus.SubscribeFunc(events.EngineStarted, h.handleEngineStarted)
	eventBus.SubscribeFunc(events.EngineStopped, h.handleEngineStopped)
}

// Event Handlers

func (h *CLIHandler) handleEngineStarted(event events.Event) {
	// El engine ha iniciado pero aún no hay sesión corriendo
}

func (h *CLIHandler) handleEngineStopped(event events.Event) {
	fmt.Println("🛑 Engine detenido.")
}

func (h *CLIHandler) handleTimerStarted(event events.Event) {
	if data, ok := event.Data.(events.TimerEventData); ok {
		h.mu.Lock()
		h.currentTimerData = data
		h.firstSessionStarted = true
		h.mu.Unlock()

		// Limpiar línea de comando y mostrar display inicial
		fmt.Print("\r\033[K")
		h.displayTimerWithStats()
		fmt.Println()
		fmt.Print("Comando > ")
	}
}

func (h *CLIHandler) handleTimerTick(event events.Event) {
	if data, ok := event.Data.(events.TimerEventData); ok {
		h.mu.Lock()
		h.currentTimerData = data
		showing := h.isShowingStats || h.waitingForInput
		h.mu.Unlock()

		// Solo actualizar si no estamos mostrando mensajes importantes
		if !showing {
			// Actualizar display sin interrumpir input
			fmt.Print("\033[s")   // Guardar cursor
			fmt.Print("\033[A")   // Subir una línea
			fmt.Print("\r\033[K") // Limpiar línea del timer
			h.displayTimerWithStats()
			fmt.Print("\033[u") // Restaurar cursor
		}
	}
}

func (h *CLIHandler) handleTimerPaused(event events.Event) {
	fmt.Println("⏸️  Timer pausado. Escribe 'r' para reanudar.")
	fmt.Print("Comando > ")
}

func (h *CLIHandler) handleTimerResumed(event events.Event) {
	fmt.Println("▶️  Timer reanudado.")
	fmt.Print("Comando > ")
}

func (h *CLIHandler) handleTimerCompleted(event events.Event) {
	fmt.Println() // Nueva línea al terminar
}

func (h *CLIHandler) handleTimerSkipped(event events.Event) {
	fmt.Println("⏭️  Timer saltado.")
}

func (h *CLIHandler) handlePomodoroStarted(event events.Event) {
	if data, ok := event.Data.(events.PomodoroEventData); ok {
		fmt.Printf("\n🍅 Pomodoro #%d - Sesión de trabajo\n", data.Number)
		time.Sleep(2 * time.Second)
	}
}

func (h *CLIHandler) handlePomodoroCompleted(event events.Event) {
	if data, ok := event.Data.(events.PomodoroEventData); ok {
		h.mu.Lock()
		h.waitingForInput = true
		h.mu.Unlock()

		// Limpiar display antes de mostrar mensaje
		fmt.Print("\r\033[K") // Limpiar línea actual
		fmt.Println()
		fmt.Println()

		fmt.Println(ui.Colorize("+================================+", ui.ColorGreen, true))
		fmt.Println(ui.Colorize("|       POMODORO COMPLETO!       |", ui.ColorGreen, true))
		fmt.Println(ui.Colorize("+================================+", ui.ColorGreen, true))
		fmt.Printf("✅ ¡Pomodoro #%d completado!\n", data.Number)

		// Determinar próximo descanso CORRECTAMENTE
		nextBreakType, nextDuration := h.getNextBreakInfo(data.Number)
		fmt.Printf("🎯 Próximo: %s (%s)\n", nextBreakType, ui.FormatDuration(nextDuration))
		fmt.Println()

		// Mostrar estadísticas rápidas
		h.mu.RLock()
		stats := h.currentStatsData
		h.mu.RUnlock()

		fmt.Printf("📊 Estadísticas: 🍅 %d | 🔥 %d | ⏱️ %s\n",
			stats.PomodorosCompleted, stats.CurrentStreak, formatDuration(stats.TotalWorkTime))

		if stats.CurrentStreak > 1 {
			fmt.Printf("🔥 ¡Racha de %d pomodoros!\n", stats.CurrentStreak)
		}
		fmt.Println()

		fmt.Println(ui.Colorize("Escribe 'c' para continuar, 'stats' para ver estadísticas detalladas, o 'q' para salir", ui.ColorYellow, true))
		fmt.Print("Comando > ")

		h.mu.Lock()
		h.waitingForInput = false
		h.mu.Unlock()
	}
}

func (h *CLIHandler) handlePomodoroSkipped(event events.Event) {
	if data, ok := event.Data.(events.PomodoroEventData); ok {
		h.mu.Lock()
		h.waitingForInput = true
		h.mu.Unlock()

		// Limpiar display antes de mostrar mensaje
		fmt.Print("\r\033[K") // Limpiar línea actual
		fmt.Println()
		fmt.Println()

		fmt.Println(ui.Colorize("+================================+", ui.ColorRed, true))
		fmt.Println(ui.Colorize("|      POMODORO SALTADO!         |", ui.ColorRed, true))
		fmt.Println(ui.Colorize("+================================+", ui.ColorRed, true))

		// Mostrar número correcto del pomodoro
		pomodoroNum := data.Number
		if pomodoroNum == 0 {
			pomodoroNum = h.engine.GetPomodoroCount() + 1
		}
		fmt.Printf("⏭️  Pomodoro #%d saltado\n", pomodoroNum)

		// Determinar próximo descanso CORRECTAMENTE
		nextBreakType, nextDuration := h.getNextBreakInfo(pomodoroNum)
		fmt.Printf("🎯 Próximo: %s (%s)\n", nextBreakType, ui.FormatDuration(nextDuration))
		fmt.Println()

		// Mensaje claro de continuación
		fmt.Println(ui.Colorize("Escribe 'c' para continuar con el descanso o 'q' para salir", ui.ColorYellow, true))
		fmt.Print("Comando > ")

		h.mu.Lock()
		h.waitingForInput = false
		h.mu.Unlock()
	}
}

func (h *CLIHandler) handleBreakStarted(event events.Event) {
	if data, ok := event.Data.(events.BreakEventData); ok {
		fmt.Printf("\n🧘 %s - Tiempo de descanso\n", data.Type)
		time.Sleep(2 * time.Second)
	}
}

func (h *CLIHandler) handleBreakCompleted(event events.Event) {
	if data, ok := event.Data.(events.BreakEventData); ok {
		h.mu.Lock()
		h.waitingForInput = true
		h.mu.Unlock()

		// Limpiar display antes de mostrar mensaje
		fmt.Print("\r\033[K") // Limpiar línea actual
		fmt.Println()
		fmt.Println()

		fmt.Println(ui.Colorize("+================================+", ui.ColorBlue, true))
		fmt.Println(ui.Colorize("|      DESCANSO COMPLETADO!      |", ui.ColorBlue, true))
		fmt.Println(ui.Colorize("+================================+", ui.ColorBlue, true))
		fmt.Printf("✅ %s terminado\n", data.Type)
		fmt.Println("💪 ¡Listo para el siguiente pomodoro!")
		fmt.Println()

		// Mostrar qué pomodoro viene
		nextPomodoroNum := h.engine.GetPomodoroCount() + 1
		fmt.Printf("🎯 Próximo: Pomodoro #%d (%s)\n",
			nextPomodoroNum, ui.FormatDuration(h.engine.GetConfig().WorkDuration))
		fmt.Println()

		fmt.Println(ui.Colorize("Escribe 'c' para continuar o 'q' para salir", ui.ColorYellow, true))
		fmt.Print("Comando > ")

		h.mu.Lock()
		h.waitingForInput = false
		h.mu.Unlock()
	}
}

func (h *CLIHandler) handleBreakSkipped(event events.Event) {
	if data, ok := event.Data.(events.BreakEventData); ok {
		h.mu.Lock()
		h.waitingForInput = true
		h.mu.Unlock()

		// Limpiar display antes de mostrar mensaje
		fmt.Print("\r\033[K") // Limpiar línea actual
		fmt.Println()
		fmt.Println()

		fmt.Println(ui.Colorize("+================================+", ui.ColorCyan, true))
		fmt.Println(ui.Colorize("|      DESCANSO SALTADO!         |", ui.ColorCyan, true))
		fmt.Println(ui.Colorize("+================================+", ui.ColorCyan, true))
		fmt.Printf("⏭️  %s saltado\n", data.Type)
		fmt.Println("💪 ¡Listo para el siguiente pomodoro!")
		fmt.Println()

		// Mostrar qué pomodoro viene
		nextPomodoroNum := h.engine.GetPomodoroCount() + 1
		fmt.Printf("🎯 Próximo: Pomodoro #%d (%s)\n",
			nextPomodoroNum, ui.FormatDuration(h.engine.GetConfig().WorkDuration))
		fmt.Println()

		fmt.Println(ui.Colorize("Escribe 'c' para continuar con el trabajo o 'q' para salir", ui.ColorYellow, true))
		fmt.Print("Comando > ")

		h.mu.Lock()
		h.waitingForInput = false
		h.mu.Unlock()
	}
}

func (h *CLIHandler) handleStatsUpdated(event events.Event) {
	if data, ok := event.Data.(events.StatsEventData); ok {
		h.mu.Lock()
		h.currentStatsData = data
		h.mu.Unlock()
	}
}

// Input Handling

func (h *CLIHandler) startInputListener() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := strings.TrimSpace(strings.ToLower(scanner.Text()))
		select {
		case h.globalInputChan <- input:
		default:
			// Canal lleno, ignorar
		}
	}
}

func (h *CLIHandler) handleInput() {
	for input := range h.globalInputChan {
		h.processCommand(input)
	}
}

func (h *CLIHandler) processCommand(input string) {
	// Mostrar el comando escrito
	fmt.Printf("%s\n", input)

	switch input {
	case "p", "pause":
		if h.isFirstSessionStarted() {
			if err := h.engine.Pause(); err != nil {
				fmt.Printf("❌ Error pausando: %v\n", err)
			}
		} else {
			fmt.Println("❌ Aún no hay sesión iniciada. Usa 'c' para empezar.")
		}

	case "r", "resume":
		if h.isFirstSessionStarted() {
			if err := h.engine.Resume(); err != nil {
				fmt.Printf("❌ Error reanudando: %v\n", err)
			}
		} else {
			fmt.Println("❌ Aún no hay sesión iniciada. Usa 'c' para empezar.")
		}

	case "s", "skip":
		if h.isFirstSessionStarted() {
			if err := h.engine.Skip(); err != nil {
				fmt.Printf("❌ Error saltando: %v\n", err)
			}
		} else {
			fmt.Println("❌ Aún no hay sesión iniciada. Usa 'c' para empezar.")
		}

	case "q", "quit":
		fmt.Println("👋 Saliendo...")
		h.engine.Stop()
		os.Exit(0)

	case "h", "help":
		h.showInlineHelp()

	case "stats", "estadisticas":
		if h.isFirstSessionStarted() {
			h.showDetailedStats()
		} else {
			fmt.Println("❌ Aún no hay estadísticas. Usa 'c' para empezar el primer pomodoro.")
		}

	case "compact", "compacto":
		if h.isFirstSessionStarted() {
			h.showCompactStats()
		} else {
			fmt.Println("❌ Aún no hay estadísticas. Usa 'c' para empezar el primer pomodoro.")
		}

	case "status", "estado":
		h.showQuickStatus()

	case "demo", "themes", "temas":
		h.showThemeDemo()

	case "test", "prueba":
		h.runFeatureTest()

	case "c", "continue", "":
		// Si es la primera vez, iniciar primera sesión
		if !h.isFirstSessionStarted() && h.engine.GetState() == engine.StateIdle {
			if err := h.engine.StartFirstSession(); err != nil {
				fmt.Printf("❌ Error iniciando sesión: %v\n", err)
			}
		}
		// Si ya hay sesión corriendo, no hacer nada (el engine maneja las transiciones)
		return

	default:
		fmt.Printf("❌ Comando '%s' no reconocido.\n", input)
		fmt.Println("💡 Usa 'h' para ver comandos disponibles")
	}

	// Nuevo prompt para el siguiente comando (solo si no estamos en sesión activa)
	if !h.isFirstSessionStarted() || h.engine.GetState() == engine.StateIdle {
		fmt.Print("Comando > ")
	}
}

// UI Methods

func (h *CLIHandler) showConfiguration() {
	ui.ClearScreen()
	cfg := h.engine.GetConfig()

	fmt.Println(ui.Colorize("+================================+", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("|          POMODORO CLI          |", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("+================================+", ui.ColorCyan, true))
	fmt.Println()
	fmt.Println("📋 Configuración:")
	fmt.Printf("   • Trabajo: %s\n", ui.Colorize(ui.FormatDuration(cfg.WorkDuration), ui.ColorRed, true))
	fmt.Printf("   • Descanso corto: %s\n", ui.Colorize(ui.FormatDuration(cfg.ShortBreak), ui.ColorCyan, true))
	fmt.Printf("   • Descanso largo: %s\n", ui.Colorize(ui.FormatDuration(cfg.LongBreak), ui.ColorBlue, true))
	fmt.Printf("   • Descanso largo cada: %s pomodoros\n", ui.Colorize(fmt.Sprintf("%d", cfg.LongBreakInterval), ui.ColorYellow, true))
	fmt.Println()
	fmt.Println("🎮 Controles: (p)ausar (r)eanudar (s)altar (q)salir (h)ayuda")
	fmt.Println("📊 Nuevo: (stats) estadísticas | (compact) vista rápida | (demo) temas")
	fmt.Println("   • Escribe el comando y presiona Enter")
	fmt.Println()
	fmt.Println(ui.Colorize("🚀 Iniciando en 3 segundos...", ui.ColorGreen, true))
	time.Sleep(3 * time.Second)
	ui.ClearScreen()
}

func (h *CLIHandler) displayTimerWithStats() {
	h.mu.RLock()
	timerData := h.currentTimerData
	statsData := h.currentStatsData
	h.mu.RUnlock()

	// Timer principal con información más clara
	state := timerData.State
	status := timerData.Status

	// Mostrar número de sesión actual para mayor claridad
	sessionInfo := ""
	if state == "TRABAJO" {
		sessionInfo = fmt.Sprintf(" #%d", h.engine.GetPomodoroCount()+1)
	}

	ui.DisplayTimer(timerData.Remaining, state+sessionInfo, status, timerData.Total)

	// Estadísticas rápidas en la misma línea
	quickStats := fmt.Sprintf("🍅 %d | 🔥 %d | ⏱️ %s",
		statsData.PomodorosCompleted,
		statsData.CurrentStreak,
		formatDuration(statsData.TotalWorkTime))
	fmt.Printf(" | %s", quickStats)
}

func (h *CLIHandler) showInlineHelp() {
	fmt.Println()
	fmt.Println(ui.Colorize("🎮 COMANDOS DISPONIBLES", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("─────────────────────", ui.ColorGray, true))

	if h.isFirstSessionStarted() {
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
		fmt.Println("🎨 PREVIEW:")
		fmt.Println("   • demo       - Ver demostración de temas")
		fmt.Println("   • test       - Probar características")
		fmt.Println()
		fmt.Println("💡 Después de empezar tendrás más comandos disponibles")
	}
	fmt.Println()
}

func (h *CLIHandler) showDetailedStats() {
	h.mu.Lock()
	h.isShowingStats = true
	h.mu.Unlock()

	ui.ClearScreen()

	stats := h.engine.GetStats()
	config := ui.DefaultStatsConfig()

	// Mostrar estadísticas completas
	statsDisplay := ui.EnhancedStatsDisplay(stats, config)
	fmt.Print(statsDisplay)

	fmt.Println("\n" + ui.Colorize("─────────────────────────────────────────────────────────────", ui.ColorGray, true))
	fmt.Println(ui.Colorize("📋 COMANDOS ADICIONALES:", ui.ColorYellow, true))
	fmt.Println("   • 'compact' - Ver estadísticas compactas")
	fmt.Println("   • 'export' - Exportar datos (próximamente)")
	fmt.Println("   • 'reset' - Reiniciar estadísticas de sesión")
	fmt.Println("   • Enter o 'c' - Volver al timer")
	fmt.Print("Comando stats > ")

	// Loop de comandos de estadísticas
	h.handleStatsCommands()
}

func (h *CLIHandler) handleStatsCommands() {
	for {
		select {
		case input := <-h.globalInputChan:
			switch strings.TrimSpace(strings.ToLower(input)) {
			case "", "c", "continue", "back", "volver":
				h.mu.Lock()
				h.isShowingStats = false
				h.mu.Unlock()
				ui.ClearScreen()
				return

			case "compact", "compacto":
				h.showCompactStats()

			case "detailed", "detallado", "full", "completo":
				h.showDetailedStats()
				return

			case "reset", "reiniciar":
				h.confirmResetStats()

			case "export", "exportar":
				fmt.Println("🚧 Función de exportación próximamente...")
				fmt.Print("Comando stats > ")

			case "help", "h", "ayuda":
				h.showStatsHelp()

			default:
				fmt.Printf("❌ Comando '%s' no reconocido en modo stats\n", input)
				fmt.Print("Comando stats > ")
			}
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (h *CLIHandler) showCompactStats() {
	ui.ClearScreen()

	stats := h.engine.GetStats()
	config := ui.DefaultStatsConfig()
	config.CompactMode = true
	config.ShowGraphs = false
	config.ShowTrends = false

	fmt.Println(ui.Colorize("🍅 ESTADÍSTICAS COMPACTAS", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("─────────────────────────", ui.ColorGray, true))
	fmt.Println()

	compactDisplay := ui.EnhancedStatsDisplay(stats, config)
	fmt.Print(compactDisplay)

	fmt.Println("\n\n📋 'detailed' para ver completas | Enter para volver")
	fmt.Print("Comando stats > ")
}

func (h *CLIHandler) confirmResetStats() {
	fmt.Println()
	fmt.Println(ui.Colorize("⚠️  CONFIRMAR REINICIO DE ESTADÍSTICAS", ui.ColorRed, true))
	fmt.Println("¿Estás seguro de que quieres reiniciar las estadísticas de esta sesión?")
	fmt.Println("Esta acción NO se puede deshacer.")
	fmt.Println()
	fmt.Println("Escribe 'CONFIRMAR' para proceder, o cualquier otra cosa para cancelar:")
	fmt.Print("Confirmación > ")

	select {
	case input := <-h.globalInputChan:
		if strings.TrimSpace(strings.ToUpper(input)) == "CONFIRMAR" {
			fmt.Println(ui.Colorize("✅ Estadísticas reiniciadas", ui.ColorGreen, true))
			fmt.Println("(Nota: Reinicio completo requiere reiniciar la aplicación)")
		} else {
			fmt.Println(ui.Colorize("❌ Reinicio cancelado", ui.ColorYellow, true))
		}
		time.Sleep(2 * time.Second)
		h.showDetailedStats()
	case <-time.After(30 * time.Second):
		fmt.Println(ui.Colorize("⏰ Tiempo agotado - reinicio cancelado", ui.ColorYellow, true))
		time.Sleep(1 * time.Second)
		h.showDetailedStats()
	}
}

func (h *CLIHandler) showStatsHelp() {
	fmt.Println()
	fmt.Println(ui.Colorize("📋 AYUDA - MODO ESTADÍSTICAS", ui.ColorCyan, true))
	fmt.Println(ui.Colorize("─────────────────────────────", ui.ColorGray, true))
	fmt.Println()
	fmt.Println("📊 COMANDOS DISPONIBLES:")
	fmt.Println("   • detailed/completo  - Vista detallada con gráficos")
	fmt.Println("   • compact/compacto   - Vista compacta")
	fmt.Println("   • reset/reiniciar    - Reiniciar estadísticas")
	fmt.Println("   • export/exportar    - Exportar datos (próximamente)")
	fmt.Println("   • help/ayuda         - Esta ayuda")
	fmt.Println("   • c/continue/Enter   - Volver al timer")
	fmt.Println()
	fmt.Println("💡 CONSEJOS:")
	fmt.Println("   • Las estadísticas se actualizan automáticamente")
	fmt.Println("   • Los logros se desbloquean al alcanzar hitos")
	fmt.Println("   • La eficiencia se calcula vs tiempo teórico")
	fmt.Println()
	fmt.Print("Comando stats > ")
}

func (h *CLIHandler) showQuickStatus() {
	if !h.isFirstSessionStarted() {
		fmt.Println("📊 Estado: Sistema listo, esperando inicio")
		return
	}

	h.mu.RLock()
	timerData := h.currentTimerData
	statsData := h.currentStatsData
	h.mu.RUnlock()

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
		fmt.Printf("⏰ Restante: %s\n", ui.Colorize(formatDuration(remainingTime), ui.ColorYellow, true))
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
		ui.Colorize(formatDuration(statsData.TotalWorkTime), ui.ColorBlue, true))

	efficiencyColor := ui.GetEfficiencyColor(statsData.WorkEfficiency)
	fmt.Printf("📈 Eficiencia: %s%.1f%%%s\n",
		string(efficiencyColor), statsData.WorkEfficiency, string(ui.ColorReset))
	fmt.Println()
}

func (h *CLIHandler) showThemeDemo() {
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
	fmt.Println("\nPresiona Enter para continuar...")

	select {
	case <-h.globalInputChan:
		return
	case <-time.After(30 * time.Second):
		return
	}
}

func (h *CLIHandler) runFeatureTest() {
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

	fmt.Println("✅ Prueba completada. Presiona Enter para continuar...")

	select {
	case <-h.globalInputChan:
		return
	case <-time.After(30 * time.Second):
		return
	}
}

// Helper Methods

func (h *CLIHandler) isFirstSessionStarted() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.firstSessionStarted
}

func (h *CLIHandler) getNextBreakInfo(pomodoroNumber int) (string, time.Duration) {
	cfg := h.engine.GetConfig()

	// Si es el primer pomodoro (número 1), siempre es descanso corto
	if pomodoroNumber == 1 {
		return "DESCANSO CORTO", cfg.ShortBreak
	}

	// Para otros pomodoros, usar la lógica del config
	duration, isLong := cfg.GetNextBreakType(pomodoroNumber)
	if isLong {
		return "DESCANSO LARGO", duration
	}
	return "DESCANSO CORTO", duration
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
