package session

import (
	"fmt"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/config"
	"github.com/kubaliski/pomodoro-cli/internal/timer"
	"github.com/kubaliski/pomodoro-cli/internal/ui"
)

// Session maneja una sesión completa de pomodoros
type Session struct {
	Config        *config.Config
	PomodoroCount int
}

// NewSession crea una nueva sesión
func NewSession(cfg *config.Config) *Session {
	return &Session{
		Config:        cfg,
		PomodoroCount: 0,
	}
}

// Run ejecuta el ciclo completo de pomodoros
func (s *Session) Run() {
	s.showConfiguration()

	for {
		s.PomodoroCount++

		// Sesión de trabajo
		fmt.Printf("\n🍅 Pomodoro #%d - Sesión de trabajo\n", s.PomodoroCount)
		if !s.runTimer(s.Config.WorkDuration, "TRABAJO") {
			return // Usuario salió con Ctrl+C
		}

		// Determinar tipo de descanso
		breakDuration, breakType := s.getBreakInfo()

		// Mostrar mensaje de completado
		s.showWorkCompleted(breakType, breakDuration)

		// Sesión de descanso
		if !s.runTimer(breakDuration, breakType) {
			return // Usuario salió con Ctrl+C
		}

		// Mostrar mensaje de descanso completado
		s.showBreakCompleted(breakType)
	}
}

func (s *Session) showConfiguration() {
	ui.ClearScreen()
	fmt.Println("+================================+")
	fmt.Println("|          POMODORO CLI          |")
	fmt.Println("+================================+")
	fmt.Println()
	fmt.Println("📋 Configuración:")
	fmt.Printf("   • Trabajo: %s\n", ui.FormatDuration(s.Config.WorkDuration))
	fmt.Printf("   • Descanso corto: %s\n", ui.FormatDuration(s.Config.ShortBreak))
	fmt.Printf("   • Descanso largo: %s\n", ui.FormatDuration(s.Config.LongBreak))
	fmt.Printf("   • Descanso largo cada: %d pomodoros\n", s.Config.LongBreakInterval)
	fmt.Println()
	fmt.Println("🚀 Iniciando en 3 segundos...")
	time.Sleep(3 * time.Second)
}

func (s *Session) getBreakInfo() (time.Duration, string) {
	if s.PomodoroCount%s.Config.LongBreakInterval == 0 {
		return s.Config.LongBreak, "DESCANSO LARGO"
	}
	return s.Config.ShortBreak, "DESCANSO"
}

func (s *Session) runTimer(duration time.Duration, state string) bool {
	pomodoroTimer := timer.NewTimer(duration)
	pomodoroTimer.Start()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for pomodoroTimer.IsRunning && !pomodoroTimer.IsFinished() {
		ui.DisplayTimer(pomodoroTimer.Remaining, state, duration)

		select {
		case <-ticker.C:
			if !pomodoroTimer.IsPaused {
				pomodoroTimer.Remaining -= time.Second
			}
		default:
			// Permitir que el programa responda a Ctrl+C
		}
	}

	return true
}

func (s *Session) showWorkCompleted(nextBreakType string, breakDuration time.Duration) {
	ui.ResetDisplay()
	ui.ClearScreen()
	fmt.Println("+================================+")
	fmt.Println("|       POMODORO COMPLETO!       |")
	fmt.Println("+================================+")
	fmt.Printf("✅ ¡Pomodoro #%d completado!\n", s.PomodoroCount)
	fmt.Printf("🎯 Próximo: %s (%s)\n", nextBreakType, ui.FormatDuration(breakDuration))
	fmt.Println("\n⏸️  Presiona Enter para continuar con el descanso...")
	fmt.Println("   (o Ctrl+C para salir)")

	fmt.Scanln()
}

func (s *Session) showBreakCompleted(breakType string) {
	ui.ResetDisplay()
	ui.ClearScreen()
	fmt.Println("+================================+")
	fmt.Println("|      DESCANSO COMPLETADO!      |")
	fmt.Println("+================================+")
	fmt.Printf("✅ %s terminado\n", breakType)
	fmt.Println("💪 ¡Listo para el siguiente pomodoro!")
	fmt.Println("\n⏸️  Presiona Enter para continuar...")
	fmt.Println("   (o Ctrl+C para salir)")

	fmt.Scanln()
}
