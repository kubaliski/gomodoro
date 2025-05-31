package notifications

import (
	"fmt"
	"time"
)

// Config contiene la configuración para el sistema de notificaciones
type Config struct {
	// Habilitación general por tipo
	SoundEnabled  bool `json:"sound_enabled"`  // Sonidos del sistema
	SystemEnabled bool `json:"system_enabled"` // Notificaciones del OS
	VisualEnabled bool `json:"visual_enabled"` // Alertas visuales en CLI

	// Habilitación por evento
	PomodoroNotifications bool `json:"pomodoro_notifications"` // Al completar pomodoros
	BreakNotifications    bool `json:"break_notifications"`    // Al completar descansos
	SystemNotifications   bool `json:"system_notifications"`   // Start/pause/resume
	EarlyAlerts           bool `json:"early_alerts"`           // Alertas tempranas (5 min)
	UrgentAlerts          bool `json:"urgent_alerts"`          // Alertas urgentes (1 min)

	// Configuración de sonidos
	SoundVolume   float64 `json:"sound_volume"`   // Volumen 0.0 - 1.0
	SoundDuration int     `json:"sound_duration"` // Duración en milisegundos
	BeepFrequency int     `json:"beep_frequency"` // Frecuencia del beep en Hz
	CustomSounds  bool    `json:"custom_sounds"`  // Usar archivos de sonido personalizados

	// Configuración de alertas de tiempo
	AlertThresholds     []int `json:"alert_thresholds"`      // Minutos para alertas [5, 2, 1]
	AlertRepeat         bool  `json:"alert_repeat"`          // Repetir alertas cada X segundos
	AlertRepeatInterval int   `json:"alert_repeat_interval"` // Intervalo de repetición en segundos

	// Configuración visual
	VisualIntensity   string `json:"visual_intensity"`    // "low", "medium", "high"
	FlashEnabled      bool   `json:"flash_enabled"`       // Parpadeo en alertas urgentes
	ColorAlerts       bool   `json:"color_alerts"`        // Cambios de color por tiempo
	ProgressBarAlerts bool   `json:"progress_bar_alerts"` // Alertas en barra de progreso

	// Configuración del sistema
	SystemPersistence int    `json:"system_persistence"` // Duración de notificaciones en segundos
	SystemActions     bool   `json:"system_actions"`     // Mostrar botones de acción
	SystemIcon        string `json:"system_icon"`        // Ruta del icono personalizado
	SystemPosition    string `json:"system_position"`    // Posición de las notificaciones

	// Configuración avanzada
	QuietHours     QuietHoursConfig `json:"quiet_hours"`     // Horarios silenciosos
	Profiles       []Profile        `json:"profiles"`        // Perfiles de configuración
	CurrentProfile string           `json:"current_profile"` // Perfil activo
}

// QuietHoursConfig configura horarios en los que se reducen las notificaciones
type QuietHoursConfig struct {
	Enabled       bool   `json:"enabled"`
	StartTime     string `json:"start_time"`     // Formato "HH:MM"
	EndTime       string `json:"end_time"`       // Formato "HH:MM"
	DisableSound  bool   `json:"disable_sound"`  // Deshabilitar sonidos
	DisableSystem bool   `json:"disable_system"` // Deshabilitar notificaciones del OS
	OnlyUrgent    bool   `json:"only_urgent"`    // Solo alertas urgentes
}

// Profile representa un perfil de configuración de notificaciones
type Profile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      Config `json:"config"`
}

// DefaultConfig retorna la configuración por defecto
func DefaultConfig() *Config {
	return &Config{
		// Habilitación por defecto
		SoundEnabled:  true,
		SystemEnabled: true,
		VisualEnabled: true,

		// Eventos habilitados
		PomodoroNotifications: true,
		BreakNotifications:    true,
		SystemNotifications:   false, // Menos intrusivo por defecto
		EarlyAlerts:           true,
		UrgentAlerts:          true,

		// Configuración de sonidos
		SoundVolume:   0.7,
		SoundDuration: 500, // 500ms
		BeepFrequency: 800, // 800Hz
		CustomSounds:  false,

		// Alertas de tiempo
		AlertThresholds:     []int{5, 2, 1}, // 5min, 2min, 1min
		AlertRepeat:         false,
		AlertRepeatInterval: 30, // 30 segundos

		// Configuración visual
		VisualIntensity:   "medium",
		FlashEnabled:      true,
		ColorAlerts:       true,
		ProgressBarAlerts: true,

		// Configuración del sistema
		SystemPersistence: 5, // 5 segundos
		SystemActions:     true,
		SystemIcon:        "", // Usar icono por defecto
		SystemPosition:    "top-right",

		// Configuración avanzada
		QuietHours: QuietHoursConfig{
			Enabled:       false,
			StartTime:     "22:00",
			EndTime:       "08:00",
			DisableSound:  true,
			DisableSystem: false,
			OnlyUrgent:    true,
		},
		Profiles:       []Profile{},
		CurrentProfile: "default",
	}
}

// Clone crea una copia profunda de la configuración
func (c *Config) Clone() *Config {
	clone := *c

	// Clonar slices
	clone.AlertThresholds = make([]int, len(c.AlertThresholds))
	copy(clone.AlertThresholds, c.AlertThresholds)

	clone.Profiles = make([]Profile, len(c.Profiles))
	copy(clone.Profiles, c.Profiles)

	return &clone
}

// Validate valida la configuración
func (c *Config) Validate() error {
	// Validar volumen
	if c.SoundVolume < 0.0 || c.SoundVolume > 1.0 {
		return fmt.Errorf("sound volume must be between 0.0 and 1.0, got %f", c.SoundVolume)
	}

	// Validar duración del sonido
	if c.SoundDuration < 100 || c.SoundDuration > 5000 {
		return fmt.Errorf("sound duration must be between 100ms and 5000ms, got %d", c.SoundDuration)
	}

	// Validar frecuencia del beep
	if c.BeepFrequency < 200 || c.BeepFrequency > 2000 {
		return fmt.Errorf("beep frequency must be between 200Hz and 2000Hz, got %d", c.BeepFrequency)
	}

	// Validar umbrales de alerta
	if len(c.AlertThresholds) == 0 {
		return fmt.Errorf("alert thresholds cannot be empty")
	}

	for _, threshold := range c.AlertThresholds {
		if threshold < 1 || threshold > 60 {
			return fmt.Errorf("alert threshold must be between 1 and 60 minutes, got %d", threshold)
		}
	}

	// Validar intervalo de repetición
	if c.AlertRepeatInterval < 5 || c.AlertRepeatInterval > 300 {
		return fmt.Errorf("alert repeat interval must be between 5 and 300 seconds, got %d", c.AlertRepeatInterval)
	}

	// Validar intensidad visual
	validIntensities := []string{"low", "medium", "high"}
	if !contains(validIntensities, c.VisualIntensity) {
		return fmt.Errorf("visual intensity must be one of %v, got %s", validIntensities, c.VisualIntensity)
	}

	// Validar persistencia del sistema
	if c.SystemPersistence < 1 || c.SystemPersistence > 30 {
		return fmt.Errorf("system persistence must be between 1 and 30 seconds, got %d", c.SystemPersistence)
	}

	// Validar posición del sistema
	validPositions := []string{"top-left", "top-right", "bottom-left", "bottom-right", "center"}
	if !contains(validPositions, c.SystemPosition) {
		return fmt.Errorf("system position must be one of %v, got %s", validPositions, c.SystemPosition)
	}

	// Validar horarios silenciosos
	if c.QuietHours.Enabled {
		if err := c.validateTimeFormat(c.QuietHours.StartTime); err != nil {
			return fmt.Errorf("invalid quiet hours start time: %w", err)
		}
		if err := c.validateTimeFormat(c.QuietHours.EndTime); err != nil {
			return fmt.Errorf("invalid quiet hours end time: %w", err)
		}
	}

	return nil
}

// validateTimeFormat valida que un string tenga formato HH:MM
func (c *Config) validateTimeFormat(timeStr string) error {
	_, err := time.Parse("15:04", timeStr)
	if err != nil {
		return fmt.Errorf("time must be in HH:MM format, got %s", timeStr)
	}
	return nil
}

// IsInQuietHours verifica si estamos en horario silencioso
func (c *Config) IsInQuietHours() bool {
	if !c.QuietHours.Enabled {
		return false
	}

	now := time.Now()
	currentTime := now.Format("15:04")

	startTime := c.QuietHours.StartTime
	endTime := c.QuietHours.EndTime

	// Caso normal: 22:00 - 08:00 (cruza medianoche)
	if startTime > endTime {
		return currentTime >= startTime || currentTime < endTime
	}

	// Caso simple: 08:00 - 22:00 (mismo día)
	return currentTime >= startTime && currentTime < endTime
}

// ApplyQuietHours aplica las restricciones de horario silencioso
func (c *Config) ApplyQuietHours() *Config {
	if !c.IsInQuietHours() {
		return c
	}

	// Crear copia modificada para horario silencioso
	quietConfig := c.Clone()

	if c.QuietHours.DisableSound {
		quietConfig.SoundEnabled = false
	}

	if c.QuietHours.DisableSystem {
		quietConfig.SystemEnabled = false
	}

	if c.QuietHours.OnlyUrgent {
		quietConfig.EarlyAlerts = false
		quietConfig.SystemNotifications = false
	}

	return quietConfig
}

// GetProfile retorna un perfil por nombre
func (c *Config) GetProfile(name string) (*Profile, error) {
	for _, profile := range c.Profiles {
		if profile.Name == name {
			return &profile, nil
		}
	}
	return nil, fmt.Errorf("profile %s not found", name)
}

// AddProfile agrega un nuevo perfil
func (c *Config) AddProfile(profile Profile) error {
	// Verificar si ya existe
	for _, p := range c.Profiles {
		if p.Name == profile.Name {
			return fmt.Errorf("profile %s already exists", profile.Name)
		}
	}

	// Validar el perfil
	if err := profile.Config.Validate(); err != nil {
		return fmt.Errorf("invalid profile config: %w", err)
	}

	c.Profiles = append(c.Profiles, profile)
	return nil
}

// RemoveProfile elimina un perfil
func (c *Config) RemoveProfile(name string) error {
	if name == "default" {
		return fmt.Errorf("cannot remove default profile")
	}

	for i, profile := range c.Profiles {
		if profile.Name == name {
			c.Profiles = append(c.Profiles[:i], c.Profiles[i+1:]...)

			// Si era el perfil actual, cambiar a default
			if c.CurrentProfile == name {
				c.CurrentProfile = "default"
			}

			return nil
		}
	}

	return fmt.Errorf("profile %s not found", name)
}

// SetActiveProfile cambia el perfil activo
func (c *Config) SetActiveProfile(name string) error {
	if name == "default" {
		c.CurrentProfile = name
		return nil
	}

	// Verificar que el perfil existe
	_, err := c.GetProfile(name)
	if err != nil {
		return err
	}

	c.CurrentProfile = name
	return nil
}

// GetActiveConfig retorna la configuración del perfil activo
func (c *Config) GetActiveConfig() *Config {
	if c.CurrentProfile == "default" {
		return c.ApplyQuietHours()
	}

	profile, err := c.GetProfile(c.CurrentProfile)
	if err != nil {
		// Fallback al default si el perfil no existe
		return c.ApplyQuietHours()
	}

	return profile.Config.ApplyQuietHours()
}

// Predefined profiles

// WorkProfile retorna un perfil optimizado para trabajo
func WorkProfile() Profile {
	config := DefaultConfig()
	config.SoundEnabled = false // Sin sonidos en el trabajo
	config.SystemEnabled = true
	config.VisualEnabled = true
	config.SystemNotifications = false
	config.EarlyAlerts = true
	config.UrgentAlerts = true
	config.VisualIntensity = "low" // Menos distractor
	config.FlashEnabled = false

	return Profile{
		Name:        "work",
		Description: "Optimizado para entornos de trabajo (sin sonidos, notificaciones discretas)",
		Config:      *config,
	}
}

// HomeProfile retorna un perfil optimizado para casa
func HomeProfile() Profile {
	config := DefaultConfig()
	config.SoundEnabled = true
	config.SystemEnabled = true
	config.VisualEnabled = true
	config.SystemNotifications = true
	config.VisualIntensity = "high"
	config.FlashEnabled = true
	config.SoundVolume = 0.8

	return Profile{
		Name:        "home",
		Description: "Perfil completo para uso en casa (todas las notificaciones activas)",
		Config:      *config,
	}
}

// FocusProfile retorna un perfil para sesiones de enfoque intenso
func FocusProfile() Profile {
	config := DefaultConfig()
	config.SoundEnabled = false
	config.SystemEnabled = false // Sin interrupciones del sistema
	config.VisualEnabled = true
	config.SystemNotifications = false
	config.EarlyAlerts = false // Solo alertas urgentes
	config.UrgentAlerts = true
	config.VisualIntensity = "medium"
	config.FlashEnabled = true

	return Profile{
		Name:        "focus",
		Description: "Mínimas interrupciones para sesiones de enfoque profundo",
		Config:      *config,
	}
}

// SilentProfile retorna un perfil completamente silencioso
func SilentProfile() Profile {
	config := DefaultConfig()
	config.SoundEnabled = false
	config.SystemEnabled = false
	config.VisualEnabled = true // Solo alertas visuales
	config.SystemNotifications = false
	config.EarlyAlerts = false
	config.UrgentAlerts = true // Solo urgentes visuales
	config.VisualIntensity = "low"
	config.FlashEnabled = false

	return Profile{
		Name:        "silent",
		Description: "Solo alertas visuales mínimas (ideal para bibliotecas o espacios silenciosos)",
		Config:      *config,
	}
}

// GetPredefinedProfiles retorna todos los perfiles predefinidos
func GetPredefinedProfiles() []Profile {
	return []Profile{
		WorkProfile(),
		HomeProfile(),
		FocusProfile(),
		SilentProfile(),
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
