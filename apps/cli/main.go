package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/handlers"
	"github.com/kubaliski/pomodoro-core/config"
	"github.com/kubaliski/pomodoro-core/engine"
)

func main() {
	// Configuración desde flags
	var (
		workDuration      = flag.Duration("work", 25*time.Minute, "Duración de la sesión de trabajo")
		shortBreak        = flag.Duration("break", 5*time.Minute, "Duración del descanso corto")
		longBreak         = flag.Duration("long", 15*time.Minute, "Duración del descanso largo")
		longBreakInterval = flag.Int("interval", 4, "Número de pomodoros antes del descanso largo")
	)
	flag.Parse()

	// Crear configuración
	cfg := &config.Config{
		WorkDuration:      *workDuration,
		ShortBreak:        *shortBreak,
		LongBreak:         *longBreak,
		LongBreakInterval: *longBreakInterval,
	}

	// Validar configuración
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Error en configuración: %v", err)
	}

	// Crear engine del core
	pomodoroEngine := engine.NewEngine(cfg)

	// Crear handler CLI que conecta el core con la UI
	cliHandler := handlers.NewCLIHandler(pomodoroEngine)

	// Ejecutar
	ctx := context.Background()
	if err := cliHandler.Run(ctx); err != nil {
		log.Fatalf("Error ejecutando CLI: %v", err)
	}
}
