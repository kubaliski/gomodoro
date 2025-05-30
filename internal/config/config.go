package config

import "time"

// Config contiene la configuracion del pomodoro
type Config struct {
	WorkDuration      time.Duration
	ShortBreak        time.Duration
	LongBreak         time.Duration
	LongBreakInterval int //Cada cuantos intervalos (int) hay descanso
}

//DefaultConfig retorna la config por defecto
func DefaultConfig() *Config {
	return &Config{
		WorkDuration:      25 * time.Minute,
		ShortBreak:        5 * time.Minute,
		LongBreak:         15 * time.Minute,
		LongBreakInterval: 4,
	}
}
