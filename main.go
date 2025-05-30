package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/config"
	"github.com/kubaliski/pomodoro-cli/internal/timer"
	"github.com/kubaliski/pomodoro-cli/internal/ui"
)

func main() {
	// Definir flags CLI
	workDuration := flag.String("work", "25m", "Duración de la sesión de trabajo (ej: 25m, 30m)")
	shortBreak := flag.String("break", "5m", "Duración del descanso corto (ej: 5m, 10m)")
	longBreak := flag.String("long-break", "15m", "Duración del descanso largo (ej: 15m, 20m)")
	help := flag.Bool("help", false, "Mostrar ayuda")

	flag.Parse()

	// Mostrar ayuda si se solicita
	if *help {
		showHelp()
		return
	}

	// Parsear duraciones
	workDur, err := time.ParseDuration(*workDuration)
	if err != nil {
		fmt.Printf("Error: Duración de trabajo inválida '%s'. Usa formato como 25m, 30m, etc.\n", *workDuration)
		return
	}

	shortDur, err := time.ParseDuration(*shortBreak)
	if err != nil {
		fmt.Printf("Error: Duración de descanso inválida '%s'. Usa formato como 5m, 10m, etc.\n", *shortBreak)
		return
	}

	longDur, err := time.ParseDuration(*longBreak)
	if err != nil {
		fmt.Printf("Error: Duración de descanso largo inválida '%s'. Usa formato como 15m, 20m, etc.\n", *longBreak)
		return
	}

	// Crear configuración personalizada
	cfg := &config.Config{
		WorkDuration:      workDur,
		ShortBreak:        shortDur,
		LongBreak:         longDur,
		LongBreakInterval: 4,
	}

	// Mostrar configuración
	showConfiguration(cfg)

	// Ejecutar ciclo completo de pomodoro
	runPomodoroSession(cfg)
}

func showConfiguration(cfg *config.Config) {
	ui.ClearScreen()
	fmt.Println("+================================+")
	fmt.Println("|          POMODORO CLI          |")
	fmt.Println("+================================+")
	fmt.Println()
	fmt.Println(" Configuración:")
	fmt.Printf("   • Trabajo: %s\n", ui.FormatDuration(cfg.WorkDuration))
	fmt.Printf("   • Descanso corto: %s\n", ui.FormatDuration(cfg.ShortBreak))
	fmt.Printf("   • Descanso largo: %s\n", ui.FormatDuration(cfg.LongBreak))
	fmt.Printf("   • Descanso largo cada: %d pomodoros\n", cfg.LongBreakInterval)
	fmt.Println()
	fmt.Println(" Iniciando en 3 segundos...")
	time.Sleep(3 * time.Second)
}

func runPomodoroSession(cfg *config.Config) {
	pomodoroCount := 0

	for {
		pomodoroCount++

		// Sesión de trabajo
		fmt.Printf("\n🍅 Pomodoro #%d - Sesión de trabajo\n", pomodoroCount)
		if !runTimer(cfg.WorkDuration, "TRABAJO") {
			return // Usuario salió con Ctrl+C
		}

		// Determinar tipo de descanso
		var breakDuration time.Duration
		var breakType string

		if pomodoroCount%cfg.LongBreakInterval == 0 {
			breakDuration = cfg.LongBreak
			breakType = "DESCANSO LARGO"
		} else {
			breakDuration = cfg.ShortBreak
			breakType = "DESCANSO"
		}

		// Mostrar mensaje de completado
		showWorkCompleted(pomodoroCount, breakType, breakDuration)

		// Sesión de descanso
		if !runTimer(breakDuration, breakType) {
			return // Usuario salió con Ctrl+C
		}

		// Mostrar mensaje de descanso completado
		showBreakCompleted(breakType)
	}
}

func runTimer(duration time.Duration, state string) bool {
	timer := timer.NewTimer(duration)
	timer.Start()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for timer.IsRunning && !timer.IsFinished() {
		ui.DisplayTimer(timer.Remaining, state, duration)

		select {
		case <-ticker.C:
			if !timer.IsPaused {
				timer.Remaining -= time.Second
			}
		default:
			// Permitir que el programa responda a Ctrl+C
		}
	}

	return true
}

func showWorkCompleted(count int, nextBreakType string, breakDuration time.Duration) {
	ui.ResetDisplay() // Resetear el display para la nueva pantalla
	ui.ClearScreen()
	fmt.Println("+================================+")
	fmt.Println("|       POMODORO COMPLETO!       |")
	fmt.Println("+================================+")
	fmt.Printf(" ¡Pomodoro #%d completado!\n", count)
	fmt.Printf(" Próximo: %s (%s)\n", nextBreakType, ui.FormatDuration(breakDuration))
	fmt.Println("\n Presiona Enter para continuar con el descanso...")
	fmt.Println("   (o Ctrl+C para salir)")

	// Esperar a que el usuario presione Enter
	fmt.Scanln()
}

func showBreakCompleted(breakType string) {
	ui.ResetDisplay() // Resetear el display para la nueva pantalla
	ui.ClearScreen()
	fmt.Println("+================================+")
	fmt.Println("|      DESCANSO COMPLETADO!      |")
	fmt.Println("+================================+")
	fmt.Printf(" %s terminado\n", breakType)
	fmt.Println(" ¡Listo para el siguiente pomodoro!")
	fmt.Println("\n  Presiona Enter para continuar...")
	fmt.Println("   (o Ctrl+C para salir)")

	// Esperar a que el usuario presione Enter
	fmt.Scanln()
}

func showHelp() {
	fmt.Println("+================================+")
	fmt.Println("|          POMODORO CLI          |")
	fmt.Println("+================================+")
	fmt.Println()
	fmt.Println("Un temporizador Pomodoro completo con ciclos automáticos.")
	fmt.Println()
	fmt.Println("Uso:")
	fmt.Println("  pomodoro [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -work string")
	fmt.Println("        Duración de la sesión de trabajo (default \"25m\")")
	fmt.Println("  -break string")
	fmt.Println("        Duración del descanso corto (default \"5m\")")
	fmt.Println("  -long-break string")
	fmt.Println("        Duración del descanso largo (default \"15m\")")
	fmt.Println("  -help")
	fmt.Println("        Mostrar esta ayuda")
	fmt.Println()
	fmt.Println("Ejemplos:")
	fmt.Println("  pomodoro                    # Configuración estándar (25m/5m/15m)")
	fmt.Println("  pomodoro -work=30m          # Sesiones de 30 minutos")
	fmt.Println("  pomodoro -work=45m -break=10m -long-break=20m")
	fmt.Println("  pomodoro -work=5s -break=3s # Para pruebas rápidas")
	fmt.Println()
	fmt.Println("Funcionamiento:")
	fmt.Println("  • Alterna automáticamente entre trabajo y descansos")
	fmt.Println("  • Descanso largo cada 4 pomodoros completados")
	fmt.Println("  • Usa Ctrl+C para salir en cualquier momento")
}
