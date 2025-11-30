package testhelpers

import (
	"time"

	"github.com/kubaliski/pomodoro-core/config"
)

// ConfigFixtures contiene configuraciones predefinidas para testing
var ConfigFixtures = struct {
	Default   *config.Config
	Fast      *config.Config
	Medium    *config.Config
	Minimal   *config.Config
	Extended  *config.Config
	Invalid   map[string]*config.Config
	EdgeCases map[string]*config.Config
}{
	// Configuración por defecto del sistema
	Default: &config.Config{
		WorkDuration:      25 * time.Minute,
		ShortBreak:        5 * time.Minute,
		LongBreak:         15 * time.Minute,
		LongBreakInterval: 4,
	},

	// Configuración optimizada para tests rápidos
	Fast: &config.Config{
		WorkDuration:      1 * time.Minute,
		ShortBreak:        1 * time.Minute,
		LongBreak:         5 * time.Minute,
		LongBreakInterval: 2,
	},

	// Configuración para tests de duración media
	Medium: &config.Config{
		WorkDuration:      3 * time.Minute,
		ShortBreak:        2 * time.Minute,
		LongBreak:         7 * time.Minute,
		LongBreakInterval: 3,
	},

	// Configuración mínima válida
	Minimal: &config.Config{
		WorkDuration:      1 * time.Minute,
		ShortBreak:        1 * time.Minute,
		LongBreak:         5 * time.Minute,
		LongBreakInterval: 2,
	},

	// Configuración con tiempos extendidos
	Extended: &config.Config{
		WorkDuration:      90 * time.Minute,
		ShortBreak:        20 * time.Minute,
		LongBreak:         45 * time.Minute,
		LongBreakInterval: 6,
	},

	// Configuraciones inválidas para testing de validación
	Invalid: map[string]*config.Config{
		"work_duration_zero": {
			WorkDuration:      0,
			ShortBreak:        5 * time.Minute,
			LongBreak:         15 * time.Minute,
			LongBreakInterval: 4,
		},
		"work_duration_negative": {
			WorkDuration:      -10 * time.Minute,
			ShortBreak:        5 * time.Minute,
			LongBreak:         15 * time.Minute,
			LongBreakInterval: 4,
		},
		"work_duration_too_long": {
			WorkDuration:      150 * time.Minute, // > 2 horas
			ShortBreak:        5 * time.Minute,
			LongBreak:         15 * time.Minute,
			LongBreakInterval: 4,
		},
		"short_break_zero": {
			WorkDuration:      25 * time.Minute,
			ShortBreak:        0,
			LongBreak:         15 * time.Minute,
			LongBreakInterval: 4,
		},
		"short_break_too_long": {
			WorkDuration:      25 * time.Minute,
			ShortBreak:        35 * time.Minute, // > 30 minutos
			LongBreak:         40 * time.Minute,
			LongBreakInterval: 4,
		},
		"long_break_too_short": {
			WorkDuration:      25 * time.Minute,
			ShortBreak:        5 * time.Minute,
			LongBreak:         3 * time.Minute, // < 5 minutos
			LongBreakInterval: 4,
		},
		"long_break_too_long": {
			WorkDuration:      25 * time.Minute,
			ShortBreak:        5 * time.Minute,
			LongBreak:         70 * time.Minute, // > 1 hora
			LongBreakInterval: 4,
		},
		"long_break_shorter_than_short": {
			WorkDuration:      25 * time.Minute,
			ShortBreak:        10 * time.Minute,
			LongBreak:         8 * time.Minute, // Menor que short break
			LongBreakInterval: 4,
		},
		"interval_too_small": {
			WorkDuration:      25 * time.Minute,
			ShortBreak:        5 * time.Minute,
			LongBreak:         15 * time.Minute,
			LongBreakInterval: 1, // < 2
		},
		"interval_too_large": {
			WorkDuration:      25 * time.Minute,
			ShortBreak:        5 * time.Minute,
			LongBreak:         15 * time.Minute,
			LongBreakInterval: 15, // > 10
		},
	},

	// Casos edge que son técnicamente válidos pero inusuales
	EdgeCases: map[string]*config.Config{
		"minimum_valid": {
			WorkDuration:      1 * time.Minute,
			ShortBreak:        1 * time.Minute,
			LongBreak:         5 * time.Minute,
			LongBreakInterval: 2,
		},
		"maximum_valid": {
			WorkDuration:      120 * time.Minute,
			ShortBreak:        30 * time.Minute,
			LongBreak:         60 * time.Minute,
			LongBreakInterval: 10,
		},
		"long_break_barely_longer": {
			WorkDuration:      25 * time.Minute,
			ShortBreak:        5 * time.Minute,
			LongBreak:         6 * time.Minute, // Solo 1 minuto más
			LongBreakInterval: 4,
		},
		"very_frequent_long_breaks": {
			WorkDuration:      25 * time.Minute,
			ShortBreak:        5 * time.Minute,
			LongBreak:         15 * time.Minute,
			LongBreakInterval: 2, // Cada 2 pomodoros
		},
		"very_infrequent_long_breaks": {
			WorkDuration:      25 * time.Minute,
			ShortBreak:        5 * time.Minute,
			LongBreak:         15 * time.Minute,
			LongBreakInterval: 10, // Cada 10 pomodoros
		},
	},
}

// TimeFixtures contiene tiempos predefinidos para testing consistente
var TimeFixtures = struct {
	BaseTime     time.Time
	StartOfDay   time.Time
	EndOfDay     time.Time
	WorkingHours time.Time
	Timezone     *time.Location
}{
	BaseTime:     time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
	StartOfDay:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	EndOfDay:     time.Date(2024, 1, 1, 23, 59, 59, 999999999, time.UTC),
	WorkingHours: time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC),
	Timezone:     time.UTC,
}

// TestScenarios define escenarios comunes de testing
var TestScenarios = struct {
	SinglePomodoro         TestScenario
	PomodoroWithShortBreak TestScenario
	TwoPomodoros           TestScenario
	FullCycleWithLongBreak TestScenario
	InterruptedSession     TestScenario
}{
	SinglePomodoro: TestScenario{
		Name:        "Single Pomodoro",
		Description: "Complete one work session",
		Config:      ConfigFixtures.Fast,
		ExpectedEvents: []string{
			"pomodoro_started",
			"timer_started",
			"timer_tick",
			"timer_completed",
			"pomodoro_completed",
		},
	},

	PomodoroWithShortBreak: TestScenario{
		Name:        "Pomodoro + Short Break",
		Description: "Complete work session followed by short break",
		Config:      ConfigFixtures.Fast,
		ExpectedEvents: []string{
			"pomodoro_started",
			"pomodoro_completed",
			"break_started",
			"break_completed",
		},
	},

	TwoPomodoros: TestScenario{
		Name:        "Two Pomodoros",
		Description: "Complete two work sessions with break in between",
		Config:      ConfigFixtures.Fast,
		ExpectedEvents: []string{
			"pomodoro_started",
			"pomodoro_completed",
			"break_started",
			"break_completed",
			"pomodoro_started",
			"pomodoro_completed",
		},
	},

	FullCycleWithLongBreak: TestScenario{
		Name:        "Full Cycle with Long Break",
		Description: "Complete enough pomodoros to trigger long break",
		Config:      ConfigFixtures.Fast,
		ExpectedEvents: []string{
			"pomodoro_started",
			"pomodoro_completed",
			"break_started",
			"break_completed",
			"pomodoro_started",
			"pomodoro_completed",
			"break_started", // Este será long break
			"break_completed",
		},
	},

	InterruptedSession: TestScenario{
		Name:        "Interrupted Session",
		Description: "Session with pause, resume, and skip operations",
		Config:      ConfigFixtures.Fast,
		ExpectedEvents: []string{
			"pomodoro_started",
			"timer_paused",
			"timer_resumed",
			"timer_skipped",
			"pomodoro_skipped",
		},
	},
}

// TestScenario define un escenario de prueba completo
type TestScenario struct {
	Name           string
	Description    string
	Config         *config.Config
	ExpectedEvents []string
	Duration       time.Duration
	Setup          func() error
	Cleanup        func() error
}

// GetTestDuration calcula la duración esperada de un escenario
func (ts *TestScenario) GetTestDuration() time.Duration {
	if ts.Duration != 0 {
		return ts.Duration
	}

	// Calcular basado en configuración y eventos esperados
	var totalDuration time.Duration
	pomodoroCount := 0
	breakCount := 0

	for _, event := range ts.ExpectedEvents {
		switch event {
		case "pomodoro_completed":
			pomodoroCount++
		case "break_completed":
			breakCount++
		}
	}

	totalDuration += time.Duration(pomodoroCount) * ts.Config.WorkDuration
	totalDuration += time.Duration(breakCount) * ts.Config.ShortBreak

	// Agregar margen para procesamiento
	return totalDuration + (100 * time.Millisecond)
}

// GetConfigByName retorna una configuración por nombre
func GetConfigByName(name string) *config.Config {
	switch name {
	case "default":
		return ConfigFixtures.Default.Clone()
	case "fast":
		return ConfigFixtures.Fast.Clone()
	case "medium":
		return ConfigFixtures.Medium.Clone()
	case "minimal":
		return ConfigFixtures.Minimal.Clone()
	case "extended":
		return ConfigFixtures.Extended.Clone()
	default:
		if cfg, exists := ConfigFixtures.Invalid[name]; exists {
			return cfg.Clone()
		}
		if cfg, exists := ConfigFixtures.EdgeCases[name]; exists {
			return cfg.Clone()
		}
		return nil
	}
}

// GetInvalidConfigNames retorna todos los nombres de configuraciones inválidas
func GetInvalidConfigNames() []string {
	names := make([]string, 0, len(ConfigFixtures.Invalid))
	for name := range ConfigFixtures.Invalid {
		names = append(names, name)
	}
	return names
}

// GetEdgeCaseConfigNames retorna todos los nombres de configuraciones edge case
func GetEdgeCaseConfigNames() []string {
	names := make([]string, 0, len(ConfigFixtures.EdgeCases))
	for name := range ConfigFixtures.EdgeCases {
		names = append(names, name)
	}
	return names
}

// GetAllValidConfigNames retorna todos los nombres de configuraciones válidas
func GetAllValidConfigNames() []string {
	return []string{"default", "fast", "medium", "minimal", "extended"}
}
