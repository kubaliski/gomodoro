package notifications

import (
	"fmt"
	"sync"
	"time"
)

// NotificationType define los tipos de notificaciones disponibles
type NotificationType string

const (
	TypeSound  NotificationType = "sound"
	TypeSystem NotificationType = "system"
	TypeVisual NotificationType = "visual"
)

// EventType define los tipos de eventos que pueden generar notificaciones
type EventType string

const (
	EventPomodoroCompleted EventType = "pomodoro_completed"
	EventBreakCompleted    EventType = "break_completed"
	EventSessionStarted    EventType = "session_started"
	EventTimerPaused       EventType = "timer_paused"
	EventTimerResumed      EventType = "timer_resumed"
	EventEarlyAlert        EventType = "early_alert"  // 5 minutos restantes
	EventUrgentAlert       EventType = "urgent_alert" // 1 minuto restante
	EventCustomAlert       EventType = "custom_alert" // Alertas personalizadas
)

// Priority define la prioridad de las notificaciones
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityNormal Priority = 2
	PriorityHigh   Priority = 3
	PriorityUrgent Priority = 4
)

// NotificationRequest representa una solicitud de notificación
type NotificationRequest struct {
	Event         EventType
	Title         string
	Message       string
	Priority      Priority
	Types         []NotificationType     // Tipos de notificación a usar
	Metadata      map[string]interface{} // Datos adicionales
	TimeRemaining time.Duration          // Para alertas de tiempo
}

// NotificationResponse representa el resultado de una notificación
type NotificationResponse struct {
	Success  bool
	Type     NotificationType
	Error    error
	Duration time.Duration // Tiempo que tomó ejecutar
}

// Notifier define la interfaz para diferentes tipos de notificadores
type Notifier interface {
	Notify(request NotificationRequest) NotificationResponse
	IsAvailable() bool
	GetType() NotificationType
	Configure(config map[string]interface{}) error
}

// Manager es el gestor central de notificaciones
type Manager struct {
	mu        sync.RWMutex
	config    *Config
	notifiers map[NotificationType]Notifier
	enabled   bool
	stats     *NotificationStats
}

// NotificationStats mantiene estadísticas de notificaciones
type NotificationStats struct {
	TotalSent    int64
	SuccessCount int64
	FailureCount int64
	LastNotified time.Time
	ByType       map[NotificationType]int64
	ByEvent      map[EventType]int64
}

// NewManager crea un nuevo manager de notificaciones
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	return &Manager{
		config:    config,
		notifiers: make(map[NotificationType]Notifier),
		enabled:   true,
		stats: &NotificationStats{
			ByType:  make(map[NotificationType]int64),
			ByEvent: make(map[EventType]int64),
		},
	}
}

// RegisterNotifier registra un notificador específico
func (m *Manager) RegisterNotifier(notifier Notifier) error {
	if notifier == nil {
		return fmt.Errorf("notifier cannot be nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	notifierType := notifier.GetType()

	// Verificar si el notificador está disponible
	if !notifier.IsAvailable() {
		return fmt.Errorf("notifier %s is not available on this system", notifierType)
	}

	// Configurar el notificador con la configuración actual
	if err := m.configureNotifier(notifier); err != nil {
		return fmt.Errorf("failed to configure notifier %s: %w", notifierType, err)
	}

	m.notifiers[notifierType] = notifier
	return nil
}

// Notify envía una notificación usando los tipos especificados
func (m *Manager) Notify(request NotificationRequest) []NotificationResponse {
	m.mu.RLock()
	enabled := m.enabled
	config := m.config
	m.mu.RUnlock()

	var responses []NotificationResponse

	// Verificar si las notificaciones están habilitadas globalmente
	if !enabled {
		return responses
	}

	// Aplicar configuración de horarios silenciosos
	activeConfig := config.ApplyQuietHours()

	// Verificar si el evento específico está habilitado
	if !m.isEventEnabledWithConfig(request.Event, activeConfig) {
		return responses
	}

	// Si no se especifican tipos, usar todos los habilitados
	if len(request.Types) == 0 {
		request.Types = m.getEnabledTypesWithConfig(activeConfig)
	}

	// Ejecutar notificaciones en paralelo
	var wg sync.WaitGroup
	responseChan := make(chan NotificationResponse, len(request.Types))

	for _, notificationType := range request.Types {
		if !m.isTypeEnabledWithConfig(notificationType, activeConfig) {
			continue
		}

		wg.Add(1)
		go func(nType NotificationType) {
			defer wg.Done()
			response := m.sendNotification(nType, request, activeConfig)
			responseChan <- response
		}(notificationType)
	}

	// Esperar a que todas las notificaciones terminen
	go func() {
		wg.Wait()
		close(responseChan)
	}()

	// Recopilar respuestas
	for response := range responseChan {
		responses = append(responses, response)
		m.updateStats(request.Event, response)
	}

	m.mu.Lock()
	m.stats.LastNotified = time.Now()
	m.mu.Unlock()

	return responses
}

// sendNotification envía una notificación de un tipo específico
func (m *Manager) sendNotification(notificationType NotificationType, request NotificationRequest, config *Config) NotificationResponse {
	m.mu.RLock()
	notifier, exists := m.notifiers[notificationType]
	m.mu.RUnlock()

	if !exists {
		return NotificationResponse{
			Success: false,
			Type:    notificationType,
			Error:   fmt.Errorf("notifier %s not registered", notificationType),
		}
	}

	// Verificar configuración específica del tipo
	if !m.isTypeEnabledWithConfig(notificationType, config) {
		return NotificationResponse{
			Success: false,
			Type:    notificationType,
			Error:   fmt.Errorf("notification type %s disabled by configuration", notificationType),
		}
	}

	// Aplicar configuración específica según el tipo y evento
	configuredRequest := m.configureRequest(request, notificationType, config)

	start := time.Now()
	response := notifier.Notify(configuredRequest)
	response.Duration = time.Since(start)
	response.Type = notificationType

	return response
}

// QuickNotify es un helper para notificaciones simples
func (m *Manager) QuickNotify(event EventType, title, message string, priority Priority) []NotificationResponse {
	return m.Notify(NotificationRequest{
		Event:    event,
		Title:    title,
		Message:  message,
		Priority: priority,
		Types:    m.getEnabledTypes(),
	})
}

// NotifyPomodoroCompleted notificación específica para pomodoro completado
func (m *Manager) NotifyPomodoroCompleted(pomodoroNumber int, nextBreakDuration time.Duration) []NotificationResponse {
	return m.Notify(NotificationRequest{
		Event:    EventPomodoroCompleted,
		Title:    "🍅 ¡Pomodoro Completado!",
		Message:  fmt.Sprintf("Pomodoro #%d terminado. Descanso de %s.", pomodoroNumber, formatDuration(nextBreakDuration)),
		Priority: PriorityHigh,
		Metadata: map[string]interface{}{
			"pomodoro_number": pomodoroNumber,
			"break_duration":  nextBreakDuration,
		},
	})
}

// NotifyBreakCompleted notificación específica para descanso completado
func (m *Manager) NotifyBreakCompleted(breakType string, nextPomodoroNumber int) []NotificationResponse {
	return m.Notify(NotificationRequest{
		Event:    EventBreakCompleted,
		Title:    "🧘 ¡Descanso Completado!",
		Message:  fmt.Sprintf("%s terminado. Listo para Pomodoro #%d.", breakType, nextPomodoroNumber),
		Priority: PriorityHigh,
		Metadata: map[string]interface{}{
			"break_type":    breakType,
			"next_pomodoro": nextPomodoroNumber,
		},
	})
}

// NotifyTimeAlert notificación de alerta de tiempo
func (m *Manager) NotifyTimeAlert(timeRemaining time.Duration, sessionType string) []NotificationResponse {
	var event EventType
	var priority Priority
	var emoji string

	if timeRemaining <= time.Minute {
		event = EventUrgentAlert
		priority = PriorityUrgent
		emoji = "🚨"
	} else if timeRemaining <= 5*time.Minute {
		event = EventEarlyAlert
		priority = PriorityHigh
		emoji = "⚠️"
	} else {
		event = EventCustomAlert
		priority = PriorityNormal
		emoji = "⏰"
	}

	return m.Notify(NotificationRequest{
		Event:         event,
		Title:         fmt.Sprintf("%s Alerta de Tiempo", emoji),
		Message:       fmt.Sprintf("%s restantes en %s", formatDuration(timeRemaining), sessionType),
		Priority:      priority,
		TimeRemaining: timeRemaining,
		Types:         []NotificationType{TypeVisual, TypeSound}, // No usar sistema para alertas frecuentes
		Metadata: map[string]interface{}{
			"time_remaining": timeRemaining,
			"session_type":   sessionType,
		},
	})
}

// NotifySessionStarted notificación de inicio de sesión
func (m *Manager) NotifySessionStarted(sessionType string, duration time.Duration) []NotificationResponse {
	return m.Notify(NotificationRequest{
		Event:    EventSessionStarted,
		Title:    "🚀 Sesión Iniciada",
		Message:  fmt.Sprintf("%s iniciado (%s)", sessionType, formatDuration(duration)),
		Priority: PriorityNormal,
		Metadata: map[string]interface{}{
			"session_type": sessionType,
			"duration":     duration,
		},
	})
}

// NotifyTimerPaused notificación de timer pausado
func (m *Manager) NotifyTimerPaused() []NotificationResponse {
	return m.Notify(NotificationRequest{
		Event:    EventTimerPaused,
		Title:    "⏸️ Timer Pausado",
		Message:  "El timer ha sido pausado",
		Priority: PriorityNormal,
	})
}

// NotifyTimerResumed notificación de timer reanudado
func (m *Manager) NotifyTimerResumed(timeRemaining time.Duration) []NotificationResponse {
	return m.Notify(NotificationRequest{
		Event:         EventTimerResumed,
		Title:         "▶️ Timer Reanudado",
		Message:       fmt.Sprintf("Timer reanudado. %s restantes", formatDuration(timeRemaining)),
		Priority:      PriorityNormal,
		TimeRemaining: timeRemaining,
	})
}

// Configuration methods

// Enable habilita o deshabilita todas las notificaciones
func (m *Manager) Enable(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = enabled
}

// IsEnabled retorna si las notificaciones están habilitadas
func (m *Manager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled
}

// UpdateConfig actualiza la configuración
func (m *Manager) UpdateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config

	// Reconfigurar todos los notificadores registrados
	for _, notifier := range m.notifiers {
		if err := m.configureNotifier(notifier); err != nil {
			// Log el error pero no fallar completamente
			fmt.Printf("Warning: failed to reconfigure notifier %s: %v\n", notifier.GetType(), err)
		}
	}

	return nil
}

// GetConfig retorna una copia de la configuración actual
func (m *Manager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.Clone()
}

// GetStats retorna las estadísticas de notificaciones
func (m *Manager) GetStats() NotificationStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Crear copia para evitar race conditions
	stats := *m.stats
	stats.ByType = make(map[NotificationType]int64)
	stats.ByEvent = make(map[EventType]int64)

	for k, v := range m.stats.ByType {
		stats.ByType[k] = v
	}
	for k, v := range m.stats.ByEvent {
		stats.ByEvent[k] = v
	}

	return stats
}

// GetRegisteredNotifiers retorna los tipos de notificadores registrados
func (m *Manager) GetRegisteredNotifiers() []NotificationType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var types []NotificationType
	for notifierType := range m.notifiers {
		types = append(types, notifierType)
	}
	return types
}

// TestNotifications prueba todos los notificadores registrados
func (m *Manager) TestNotifications() map[NotificationType]NotificationResponse {
	results := make(map[NotificationType]NotificationResponse)

	request := NotificationRequest{
		Event:    EventCustomAlert,
		Title:    "🧪 Test de Notificaciones",
		Message:  "Esta es una notificación de prueba",
		Priority: PriorityNormal,
	}

	m.mu.RLock()
	notifiers := make(map[NotificationType]Notifier)
	for k, v := range m.notifiers {
		notifiers[k] = v
	}
	config := m.config
	m.mu.RUnlock()

	for notifierType := range notifiers {
		if m.isTypeEnabledWithConfig(notifierType, config) {
			response := m.sendNotification(notifierType, request, config)
			results[notifierType] = response
		}
	}

	return results
}

// Private helper methods

// isEventEnabledWithConfig verifica si un evento específico está habilitado
func (m *Manager) isEventEnabledWithConfig(event EventType, config *Config) bool {
	switch event {
	case EventPomodoroCompleted:
		return config.PomodoroNotifications
	case EventBreakCompleted:
		return config.BreakNotifications
	case EventEarlyAlert:
		return config.EarlyAlerts
	case EventUrgentAlert:
		return config.UrgentAlerts
	case EventSessionStarted, EventTimerPaused, EventTimerResumed:
		return config.SystemNotifications
	default:
		return true
	}
}

// isTypeEnabledWithConfig verifica si un tipo de notificación está habilitado
func (m *Manager) isTypeEnabledWithConfig(notificationType NotificationType, config *Config) bool {
	switch notificationType {
	case TypeSound:
		return config.SoundEnabled
	case TypeSystem:
		return config.SystemEnabled
	case TypeVisual:
		return config.VisualEnabled
	default:
		return false
	}
}

// getEnabledTypes retorna los tipos de notificación habilitados (método legacy)
func (m *Manager) getEnabledTypes() []NotificationType {
	m.mu.RLock()
	config := m.config
	m.mu.RUnlock()

	return m.getEnabledTypesWithConfig(config)
}

// getEnabledTypesWithConfig retorna los tipos de notificación habilitados
func (m *Manager) getEnabledTypesWithConfig(config *Config) []NotificationType {
	var types []NotificationType

	if config.SoundEnabled {
		types = append(types, TypeSound)
	}
	if config.SystemEnabled {
		types = append(types, TypeSystem)
	}
	if config.VisualEnabled {
		types = append(types, TypeVisual)
	}

	return types
}

// configureNotifier configura un notificador con la configuración actual
func (m *Manager) configureNotifier(notifier Notifier) error {
	notifierConfig := make(map[string]interface{})

	switch notifier.GetType() {
	case TypeSound:
		notifierConfig["volume"] = m.config.SoundVolume
		notifierConfig["duration"] = m.config.SoundDuration
		notifierConfig["frequency"] = m.config.BeepFrequency
		notifierConfig["custom_sounds"] = m.config.CustomSounds

	case TypeSystem:
		notifierConfig["persistence"] = m.config.SystemPersistence
		notifierConfig["actions"] = m.config.SystemActions
		notifierConfig["icon"] = m.config.SystemIcon
		notifierConfig["position"] = m.config.SystemPosition

	case TypeVisual:
		notifierConfig["intensity"] = m.config.VisualIntensity
		notifierConfig["flash_enabled"] = m.config.FlashEnabled
		notifierConfig["color_alerts"] = m.config.ColorAlerts
		notifierConfig["progress_bar_alerts"] = m.config.ProgressBarAlerts
	}

	return notifier.Configure(notifierConfig)
}

// configureRequest modifica la request según el tipo de notificador y configuración
func (m *Manager) configureRequest(request NotificationRequest, notificationType NotificationType, config *Config) NotificationRequest {
	configured := request

	// Aplicar repetición de alertas si está habilitada
	if config.AlertRepeat && (request.Event == EventEarlyAlert || request.Event == EventUrgentAlert) {
		if configured.Metadata == nil {
			configured.Metadata = make(map[string]interface{})
		}
		configured.Metadata["repeat_interval"] = config.AlertRepeatInterval
	}

	// Configurar intensidad visual según el tipo
	if notificationType == TypeVisual {
		if configured.Metadata == nil {
			configured.Metadata = make(map[string]interface{})
		}
		configured.Metadata["visual_intensity"] = config.VisualIntensity
		configured.Metadata["flash_enabled"] = config.FlashEnabled
	}

	return configured
}

// updateStats actualiza las estadísticas
func (m *Manager) updateStats(event EventType, response NotificationResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.TotalSent++
	if response.Success {
		m.stats.SuccessCount++
	} else {
		m.stats.FailureCount++
	}

	m.stats.ByType[response.Type]++
	m.stats.ByEvent[event]++
}

// ResetStats reinicia las estadísticas
func (m *Manager) ResetStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats = &NotificationStats{
		ByType:  make(map[NotificationType]int64),
		ByEvent: make(map[EventType]int64),
	}
}

// PrintStats imprime las estadísticas en formato legible
func (m *Manager) PrintStats() {
	stats := m.GetStats()

	fmt.Println("📊 Estadísticas de Notificaciones:")
	fmt.Printf("   Total enviadas: %d\n", stats.TotalSent)
	fmt.Printf("   Exitosas: %d (%.1f%%)\n", stats.SuccessCount,
		float64(stats.SuccessCount)/float64(stats.TotalSent)*100)
	fmt.Printf("   Fallidas: %d (%.1f%%)\n", stats.FailureCount,
		float64(stats.FailureCount)/float64(stats.TotalSent)*100)

	if !stats.LastNotified.IsZero() {
		fmt.Printf("   Última notificación: %s\n", stats.LastNotified.Format("15:04:05"))
	}

	if len(stats.ByType) > 0 {
		fmt.Println("\n   Por tipo:")
		for notType, count := range stats.ByType {
			fmt.Printf("     %s: %d\n", notType, count)
		}
	}

	if len(stats.ByEvent) > 0 {
		fmt.Println("\n   Por evento:")
		for event, count := range stats.ByEvent {
			fmt.Printf("     %s: %d\n", event, count)
		}
	}
}

// formatDuration formatea una duración de manera legible
func formatDuration(d time.Duration) string {
	if d >= time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}

	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	return fmt.Sprintf("%ds", seconds)
}
