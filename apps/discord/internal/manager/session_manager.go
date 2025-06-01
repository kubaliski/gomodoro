package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kubaliski/pomodoro-core/config"
	"github.com/kubaliski/pomodoro-core/engine"
	"github.com/kubaliski/pomodoro-core/events"
)

// UserSession representa una sesión de pomodoro para un usuario específico
type UserSession struct {
	UserID    string
	ChannelID string
	Engine    engine.EngineInterface
	Config    *config.Config
	StartTime time.Time
	Active    bool
}

// SessionManager maneja múltiples sesiones de usuarios
type SessionManager struct {
	mu            sync.RWMutex
	sessions      map[string]*UserSession // userID -> session
	defaultConfig *config.Config
	eventHandlers map[string]EventHandlerFunc
}

// EventHandlerFunc maneja eventos de Discord
type EventHandlerFunc func(userID, channelID string, event events.Event)

// NewSessionManager crea un nuevo manager de sesiones
func NewSessionManager(defaultConfig *config.Config) *SessionManager {
	return &SessionManager{
		sessions:      make(map[string]*UserSession),
		defaultConfig: defaultConfig.Clone(),
		eventHandlers: make(map[string]EventHandlerFunc),
	}
}

// StartSession inicia una nueva sesión para un usuario
func (sm *SessionManager) StartSession(userID, channelID string, customConfig *config.Config) (*UserSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Verificar si ya existe una sesión activa
	if session, exists := sm.sessions[userID]; exists && session.Active {
		return nil, fmt.Errorf("user already has an active pomodoro session")
	}

	// Usar configuración custom o por defecto
	cfg := sm.defaultConfig
	if customConfig != nil {
		cfg = customConfig
	}

	// Crear nueva engine
	pomodoroEngine := engine.NewEngine(cfg.Clone())

	// Crear sesión
	session := &UserSession{
		UserID:    userID,
		ChannelID: channelID,
		Engine:    pomodoroEngine,
		Config:    cfg.Clone(),
		StartTime: time.Now(),
		Active:    true,
	}

	// Configurar event handlers para esta sesión
	sm.setupSessionEventHandlers(session)

	// Iniciar engine
	ctx := context.Background()
	if err := session.Engine.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start pomodoro engine: %w", err)
	}

	// Iniciar primera sesión
	if err := session.Engine.StartFirstSession(); err != nil {
		session.Engine.Stop()
		return nil, fmt.Errorf("failed to start first session: %w", err)
	}

	// Guardar sesión
	sm.sessions[userID] = session

	return session, nil
}

// StopSession detiene la sesión de un usuario
func (sm *SessionManager) StopSession(userID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[userID]
	if !exists || !session.Active {
		return fmt.Errorf("no active session found for user")
	}

	session.Engine.Stop()
	session.Active = false
	delete(sm.sessions, userID)

	return nil
}

// GetSession obtiene la sesión de un usuario
func (sm *SessionManager) GetSession(userID string) (*UserSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[userID]
	if !exists || !session.Active {
		return nil, fmt.Errorf("no active session found for user")
	}

	return session, nil
}

// GetAllActiveSessions retorna todas las sesiones activas
func (sm *SessionManager) GetAllActiveSessions() map[string]*UserSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	activeSessions := make(map[string]*UserSession)
	for userID, session := range sm.sessions {
		if session.Active {
			activeSessions[userID] = session
		}
	}

	return activeSessions
}

// PauseSession pausa la sesión de un usuario
func (sm *SessionManager) PauseSession(userID string) error {
	session, err := sm.GetSession(userID)
	if err != nil {
		return err
	}
	return session.Engine.Pause()
}

// ResumeSession reanuda la sesión de un usuario
func (sm *SessionManager) ResumeSession(userID string) error {
	session, err := sm.GetSession(userID)
	if err != nil {
		return err
	}
	return session.Engine.Resume()
}

// SkipSession salta la sesión actual de un usuario
func (sm *SessionManager) SkipSession(userID string) error {
	session, err := sm.GetSession(userID)
	if err != nil {
		return err
	}
	return session.Engine.Skip()
}

// RegisterEventHandler registra un handler para eventos de Discord
func (sm *SessionManager) RegisterEventHandler(eventType string, handler EventHandlerFunc) {
	sm.eventHandlers[eventType] = handler
}

// setupSessionEventHandlers configura los event handlers para una sesión
func (sm *SessionManager) setupSessionEventHandlers(session *UserSession) {
	eventBus := session.Engine.GetEventBus()

	// Handler para eventos de pomodoro completado
	eventBus.SubscribeFunc(events.PomodoroCompleted, func(event events.Event) {
		if handler, exists := sm.eventHandlers["pomodoro_completed"]; exists {
			handler(session.UserID, session.ChannelID, event)
		}
	})

	// Handler para eventos de break completado
	eventBus.SubscribeFunc(events.BreakCompleted, func(event events.Event) {
		if handler, exists := sm.eventHandlers["break_completed"]; exists {
			handler(session.UserID, session.ChannelID, event)
		}
	})

	// Handler para eventos de pomodoro iniciado
	eventBus.SubscribeFunc(events.PomodoroStarted, func(event events.Event) {
		if handler, exists := sm.eventHandlers["pomodoro_started"]; exists {
			handler(session.UserID, session.ChannelID, event)
		}
	})

	// Handler para eventos de break iniciado
	eventBus.SubscribeFunc(events.BreakStarted, func(event events.Event) {
		if handler, exists := sm.eventHandlers["break_started"]; exists {
			handler(session.UserID, session.ChannelID, event)
		}
	})

	// Handler para eventos de tick (notificar cada 5 minutos)
	lastNotified := -1
	eventBus.SubscribeFunc(events.TimerTick, func(event events.Event) {
		if data, ok := event.Data.(events.TimerEventData); ok {
			currentMinute := int(data.Remaining.Minutes())
			// Notificar en minutos específicos: 10, 5, 1
			if (currentMinute == 10 || currentMinute == 5 || currentMinute == 1) && currentMinute != lastNotified {
				lastNotified = currentMinute
				if handler, exists := sm.eventHandlers["timer_reminder"]; exists {
					handler(session.UserID, session.ChannelID, event)
				}
			}
		}
	})
}

// GetSessionStats obtiene las estadísticas de la sesión de un usuario
func (sm *SessionManager) GetSessionStats(userID string) (interface{}, error) {
	session, err := sm.GetSession(userID)
	if err != nil {
		return nil, err
	}

	return session.Engine.GetStats().GetSnapshot(), nil
}

// CleanupInactiveSessions limpia sesiones inactivas
func (sm *SessionManager) CleanupInactiveSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for userID, session := range sm.sessions {
		if !session.Active || !session.Engine.IsRunning() {
			session.Engine.Stop()
			delete(sm.sessions, userID)
		}
	}
}
