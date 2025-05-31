package handlers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/notifications"
	"github.com/kubaliski/pomodoro-core/engine"
	"github.com/kubaliski/pomodoro-core/events"
)

// CLIHandler maneja la interfaz CLI conectando el core con la UI
type CLIHandler struct {
	// Core components
	engine              engine.EngineInterface
	notificationManager *notifications.Manager

	// Estado de la UI
	currentTimerData    events.TimerEventData
	currentStatsData    events.StatsEventData
	isShowingStats      bool
	waitingForInput     bool
	firstSessionStarted bool

	// Control de concurrencia
	mu sync.RWMutex

	// Control de alertas para evitar spam
	lastAlertTime   time.Time
	lastAlertMinute int

	// Sub-handlers
	eventHandler     *EventHandler
	commandProcessor *CommandProcessor
	notificationCmds *NotificationCommands
	statsCmds        *StatsCommands
	uiHelpers        *UIHelpers
	inputManager     *InputManager
}

// NewCLIHandler crea un nuevo handler CLI
func NewCLIHandler(eng engine.EngineInterface) *CLIHandler {
	// Crear configuración de notificaciones
	notifConfig := notifications.DefaultConfig()
	notifManager := notifications.NewManager(notifConfig)

	// Registrar notificador de sonido
	soundNotifier := notifications.NewSoundNotifier()
	if err := notifManager.RegisterNotifier(soundNotifier); err != nil {
		fmt.Printf("⚠️ Warning: Sound notifications not available: %v\n", err)
	} else {
		fmt.Println("🔊 Sistema de notificaciones de sonido activado")
	}

	handler := &CLIHandler{
		engine:              eng,
		notificationManager: notifManager,
		firstSessionStarted: false,
		lastAlertMinute:     -1,
	}

	// Inicializar sub-componentes
	handler.inputManager = NewInputManager()
	handler.eventHandler = NewEventHandler(handler)
	handler.commandProcessor = NewCommandProcessor(handler)
	handler.notificationCmds = NewNotificationCommands(handler)
	handler.statsCmds = NewStatsCommands(handler)
	handler.uiHelpers = NewUIHelpers(handler)

	// Configurar eventos
	handler.eventHandler.SetupEventHandlers(eng.GetEventBus())

	// Iniciar listener de input
	go handler.inputManager.StartListener()

	return handler
}

// Run ejecuta la interfaz CLI
func (h *CLIHandler) Run(ctx context.Context) error {
	h.uiHelpers.ShowConfiguration()

	if err := h.engine.Start(ctx); err != nil {
		return fmt.Errorf("error starting engine: %w", err)
	}

	h.uiHelpers.ShowInitialPrompt()
	h.inputManager.HandleInput(h.commandProcessor.ProcessCommand)
	return nil
}

// Getters para acceso a componentes internos
func (h *CLIHandler) GetEngine() engine.EngineInterface              { return h.engine }
func (h *CLIHandler) GetNotificationManager() *notifications.Manager { return h.notificationManager }
func (h *CLIHandler) GetInputManager() *InputManager                 { return h.inputManager }
func (h *CLIHandler) GetEventHandler() *EventHandler                 { return h.eventHandler }
func (h *CLIHandler) GetCommandProcessor() *CommandProcessor         { return h.commandProcessor }
func (h *CLIHandler) GetNotificationCommands() *NotificationCommands { return h.notificationCmds }
func (h *CLIHandler) GetStatsCommands() *StatsCommands               { return h.statsCmds }
func (h *CLIHandler) GetUIHelpers() *UIHelpers                       { return h.uiHelpers }

// Thread-safe getters para estado
func (h *CLIHandler) IsFirstSessionStarted() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.firstSessionStarted
}

func (h *CLIHandler) GetCurrentTimerData() events.TimerEventData {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.currentTimerData
}

func (h *CLIHandler) GetCurrentStatsData() events.StatsEventData {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.currentStatsData
}

func (h *CLIHandler) IsShowingStats() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isShowingStats
}

func (h *CLIHandler) IsWaitingForInput() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.waitingForInput
}

func (h *CLIHandler) GetLastAlertMinute() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastAlertMinute
}

// Thread-safe setters
func (h *CLIHandler) SetFirstSessionStarted(started bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.firstSessionStarted = started
}

func (h *CLIHandler) SetCurrentTimerData(data events.TimerEventData) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.currentTimerData = data
}

func (h *CLIHandler) SetCurrentStatsData(data events.StatsEventData) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.currentStatsData = data
}

func (h *CLIHandler) SetWaitingForInput(waiting bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.waitingForInput = waiting
}

func (h *CLIHandler) SetShowingStats(showing bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.isShowingStats = showing
}

func (h *CLIHandler) UpdateLastAlert(minute int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastAlertMinute = minute
	h.lastAlertTime = time.Now()
}

// Helper method para obtener información del próximo descanso
func (h *CLIHandler) GetNextBreakInfo(pomodoroNumber int) (string, time.Duration) {
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

// Helper method para formatear duración
func FormatDuration(d time.Duration) string {
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

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
