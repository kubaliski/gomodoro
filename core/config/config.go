package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config contiene la configuración del pomodoro con validación
type Config struct {
	WorkDuration      time.Duration `json:"work_duration"`
	ShortBreak        time.Duration `json:"short_break"`
	LongBreak         time.Duration `json:"long_break"`
	LongBreakInterval int           `json:"long_break_interval"`
}

// ValidationError representa un error de validación de configuración
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
}

// DefaultConfig retorna la configuración por defecto
func DefaultConfig() *Config {
	return &Config{
		WorkDuration:      25 * time.Minute,
		ShortBreak:        5 * time.Minute,
		LongBreak:         15 * time.Minute,
		LongBreakInterval: 4,
	}
}

// Validate valida que la configuración sea correcta
func (c *Config) Validate() error {
	if c.WorkDuration < 1*time.Minute {
		return ValidationError{
			Field:   "WorkDuration",
			Message: "must be at least 1 minute",
		}
	}

	if c.WorkDuration > 120*time.Minute {
		return ValidationError{
			Field:   "WorkDuration",
			Message: "must be less than 2 hours",
		}
	}

	if c.ShortBreak < 1*time.Minute {
		return ValidationError{
			Field:   "ShortBreak",
			Message: "must be at least 1 minute",
		}
	}

	if c.ShortBreak > 30*time.Minute {
		return ValidationError{
			Field:   "ShortBreak",
			Message: "must be less than 30 minutes",
		}
	}

	if c.LongBreak < 5*time.Minute {
		return ValidationError{
			Field:   "LongBreak",
			Message: "must be at least 5 minutes",
		}
	}

	if c.LongBreak > 60*time.Minute {
		return ValidationError{
			Field:   "LongBreak",
			Message: "must be less than 1 hour",
		}
	}

	if c.LongBreakInterval < 2 {
		return ValidationError{
			Field:   "LongBreakInterval",
			Message: "must be at least 2",
		}
	}

	if c.LongBreakInterval > 10 {
		return ValidationError{
			Field:   "LongBreakInterval",
			Message: "must be less than 10",
		}
	}

	// Validación lógica: descanso largo debe ser mayor que el corto
	if c.LongBreak <= c.ShortBreak {
		return ValidationError{
			Field:   "LongBreak",
			Message: "must be longer than short break",
		}
	}

	return nil
}

// LoadFromFile carga la configuración desde un archivo JSON
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// SaveToFile guarda la configuración a un archivo JSON
func (c *Config) SaveToFile(path string) error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("cannot save invalid configuration: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Clone crea una copia profunda de la configuración
func (c *Config) Clone() *Config {
	return &Config{
		WorkDuration:      c.WorkDuration,
		ShortBreak:        c.ShortBreak,
		LongBreak:         c.LongBreak,
		LongBreakInterval: c.LongBreakInterval,
	}
}

// String retorna una representación legible de la configuración
func (c *Config) String() string {
	return fmt.Sprintf("Config{Work: %v, Short: %v, Long: %v, Interval: %d}",
		c.WorkDuration, c.ShortBreak, c.LongBreak, c.LongBreakInterval)
}

// FormatDuration convierte duración a formato legible
func FormatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// GetNextBreakType determina el tipo de descanso basado en el número de pomodoro
func (c *Config) GetNextBreakType(pomodoroNumber int) (duration time.Duration, isLong bool) {
	if pomodoroNumber%c.LongBreakInterval == 0 {
		return c.LongBreak, true
	}
	return c.ShortBreak, false
}
