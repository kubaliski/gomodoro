package main

import (
	"fmt"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/config"
	"github.com/kubaliski/pomodoro-cli/internal/timer"
	"github.com/kubaliski/pomodoro-cli/internal/ui"
)

func main() {
	// Cargar configuración por defecto
	cfg := config.DefaultConfig()

	// Crear timer de trabajo (25 minutos)
	pomodoroTimer := timer.NewTimer(cfg.WorkDuration)

	fmt.Println("🍅 Pomodoro CLI - Sesión de trabajo")
	fmt.Println("Presiona Ctrl+C para salir")
	fmt.Println()

	// Iniciar el timer
	pomodoroTimer.Start()

	// Loop principal - actualiza cada segundo
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for pomodoroTimer.IsRunning && !pomodoroTimer.IsFinished() {
		// Mostrar el estado actual
		ui.DisplayTimer(pomodoroTimer.Remaining, "TRABAJO")

		// Esperar el siguiente tick
		<-ticker.C

		// Reducir el tiempo restante
		if !pomodoroTimer.IsPaused {
			pomodoroTimer.Remaining -= time.Second
		}
	}

	// Timer terminado
	ui.ClearScreen()
	fmt.Println("+================================+")
	fmt.Println("|       POMODORO COMPLETO!       |")
	fmt.Println("|     ¡Hora de descansar!        |")
	fmt.Println("+================================+")
	fmt.Println(">> ¡Excelente trabajo!")
	fmt.Println(">> Tomate un merecido descanso.")
}
