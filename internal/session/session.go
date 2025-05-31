package session

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/config"
	"github.com/kubaliski/pomodoro-cli/internal/timer"
	"github.com/kubaliski/pomodoro-cli/internal/ui"
)

// Session maneja una sesión completa de pomodoros
type Session struct {
	Config        *config.Config
	PomodoroCount int
	inputReader   *bufio.Reader
}

// NewSession crea una nueva sesión
func NewSession(cfg *config.Config) *Session {
	return &Session{
		Config:        cfg,
		PomodoroCount: 0,
		inputReader:   bufio.NewReader(os.Stdin),
	}
}

// Run ejecuta el ciclo completo de pomodoros
func (s *Session) Run() {
	s.showConfiguration()

	for {
		s.PomodoroCount++

		// Sesión de trabajo
		fmt.Printf("\n🍅 Pomodoro #%d - Sesión de trabajo\n", s.PomodoroCount)
		time.Sleep(2 * time.Second)

		result := s.runTimerWithControls(s.Config.WorkDuration, "TRABAJO")
		if result == TimerResultQuit {
			fmt.Println("\n👋 ¡Hasta luego! Buen trabajo.")
			return
		}

		// Determinar tipo de descanso
		breakDuration, breakType := s.getBreakInfo()

		// Mostrar mensaje de completado
		if result == TimerResultCompleted {
			s.showWorkCompleted(breakType, breakDuration)
		} else if result == TimerResultSkipped {
			s.showWorkSkipped(breakType, breakDuration)
		}

		// Reiniciar el display para la siguiente sesión
		ui.ResetDisplay()

		// Sesión de descanso
		result = s.runTimerWithControls(breakDuration, breakType)
		if result == TimerResultQuit {
			fmt.Println("\n👋 ¡Hasta luego! Buen trabajo.")
			return
		}

		// Mostrar mensaje de descanso completado
		if result == TimerResultCompleted {
			s.showBreakCompleted(breakType)
		} else if result == TimerResultSkipped {
			s.showBreakSkipped(breakType)
		}

		// Reiniciar el display para la siguiente sesión
		ui.ResetDisplay()
	}
}

// TimerResult representa el resultado de una sesión de timer
type TimerResult int

const (
	TimerResultCompleted TimerResult = iota
	TimerResultSkipped
	TimerResultQuit
)

func (s *Session) runTimerWithControls(duration time.Duration, state string) TimerResult {
	pomodoroTimer := timer.NewTimer(duration)
	pomodoroTimer.Start()

	// Canal para comandos
	commandChan := make(chan string, 10)

	// Variable para controlar si el usuario está escribiendo
	userTyping := false

	// Goroutine para leer input
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			userTyping = true // Usuario terminó de escribir
			input := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if input != "" {
				select {
				case commandChan <- input:
				default:
					// Canal lleno, ignorar
				}
			}
			userTyping = false // Reset después de procesar
		}
	}()

	// Ticker para actualizar cada segundo
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Variable para control de display
	lastUpdate := time.Now()

	// Mostrar display inicial
	ui.DisplayTimer(pomodoroTimer.Remaining, state, pomodoroTimer.GetStatus(), duration)
	fmt.Print(" > ") // Prompt para comandos

	for pomodoroTimer.IsRunning && !pomodoroTimer.IsFinished() {
		select {
		case <-ticker.C:
			// Actualizar timer
			pomodoroTimer.Tick()

			// Solo actualizar display si el usuario NO está escribiendo
			if !userTyping && time.Since(lastUpdate) >= 500*time.Millisecond {
				// Limpiar línea completamente y redibujar
				fmt.Print("\r\033[K")
				ui.DisplayTimer(pomodoroTimer.Remaining, state, pomodoroTimer.GetStatus(), duration)
				fmt.Print(" > ") // Prompt para comandos
				lastUpdate = time.Now()
			}

		case input := <-commandChan:
			// Limpiar línea antes de mostrar feedback
			fmt.Print("\r\033[K")

			// Procesar comando
			switch input {
			case "p", "pause":
				if !pomodoroTimer.IsPaused {
					pomodoroTimer.Pause()
					fmt.Println("⏸️  Timer pausado. Escribe 'r' para reanudar.")
				} else {
					fmt.Println("⏸️  Timer ya está pausado.")
				}

			case "r", "resume":
				if pomodoroTimer.IsPaused {
					pomodoroTimer.Resume()
					fmt.Println("▶️  Timer reanudado.")
				} else {
					fmt.Println("▶️  Timer ya está corriendo.")
				}

			case "s", "skip":
				pomodoroTimer.Skip()
				fmt.Println("⏭️  Timer saltado.")
				return TimerResultSkipped

			case "q", "quit":
				fmt.Println("👋 Saliendo...")
				return TimerResultQuit

			case "h", "help":
				s.showHelp()

			default:
				fmt.Printf("❌ Comando '%s' no reconocido. Usa: p, r, s, q, h\n", input)
			}

			// Actualizar display después del comando
			ui.DisplayTimer(pomodoroTimer.Remaining, state, pomodoroTimer.GetStatus(), duration)
			fmt.Print(" > ") // Restaurar prompt

		default:
			// No bloquear si no hay comandos ni tick
		}

		// Verificar si fue saltado
		if pomodoroTimer.IsSkipped {
			return TimerResultSkipped
		}
	}

	return TimerResultCompleted
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
	fmt.Println("🎮 Controles interactivos:")
	fmt.Println("   • (p) pausar  • (r) reanudar  • (s) saltar  • (q) salir  • (h) ayuda")
	fmt.Println("   • Escribe el comando y presiona Enter")
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

func (s *Session) showWorkCompleted(nextBreakType string, breakDuration time.Duration) {
	fmt.Print("\r\033[K") // Limpiar línea actual
	fmt.Println()
	fmt.Println("+================================+")
	fmt.Println("|       POMODORO COMPLETO!       |")
	fmt.Println("+================================+")
	fmt.Printf("✅ ¡Pomodoro #%d completado!\n", s.PomodoroCount)
	fmt.Printf("🎯 Próximo: %s (%s)\n", nextBreakType, ui.FormatDuration(breakDuration))
	fmt.Println()
	fmt.Println("⏸️  Escribe 'c' + Enter para continuar con el descanso...")
	fmt.Println("   o 'q' + Enter para salir")
	fmt.Print(" > ")

	// Usar nuestro sistema de input unificado
	s.waitForContinue()
}

func (s *Session) showWorkSkipped(nextBreakType string, breakDuration time.Duration) {
	fmt.Print("\r\033[K") // Limpiar línea actual
	fmt.Println()
	fmt.Println("+================================+")
	fmt.Println("|      POMODORO SALTADO!         |")
	fmt.Println("+================================+")
	fmt.Printf("⏭️  Pomodoro #%d saltado\n", s.PomodoroCount)
	fmt.Printf("🎯 Próximo: %s (%s)\n", nextBreakType, ui.FormatDuration(breakDuration))
	fmt.Println()
	fmt.Println("⏸️  Escribe 'c' + Enter para continuar con el descanso...")
	fmt.Println("   o 'q' + Enter para salir")
	fmt.Print(" > ")

	s.waitForContinue()
}

func (s *Session) showBreakCompleted(breakType string) {
	fmt.Print("\r\033[K") // Limpiar línea actual
	fmt.Println()
	fmt.Println("+================================+")
	fmt.Println("|      DESCANSO COMPLETADO!      |")
	fmt.Println("+================================+")
	fmt.Printf("✅ %s terminado\n", breakType)
	fmt.Println("💪 ¡Listo para el siguiente pomodoro!")
	fmt.Println()
	fmt.Println("⏸️  Escribe 'c' + Enter para continuar...")
	fmt.Println("   o 'q' + Enter para salir")
	fmt.Print(" > ")

	s.waitForContinue()
}

func (s *Session) showBreakSkipped(breakType string) {
	fmt.Print("\r\033[K") // Limpiar línea actual
	fmt.Println()
	fmt.Println("+================================+")
	fmt.Println("|      DESCANSO SALTADO!         |")
	fmt.Println("+================================+")
	fmt.Printf("⏭️  %s saltado\n", breakType)
	fmt.Println("💪 ¡Listo para el siguiente pomodoro!")
	fmt.Println()
	fmt.Println("⏸️  Escribe 'c' + Enter para continuar...")
	fmt.Println("   o 'q' + Enter para salir")
	fmt.Print(" > ")

	s.waitForContinue()
}

// waitForContinue espera a que el usuario escriba 'c' para continuar
func (s *Session) waitForContinue() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if scanner.Scan() {
			input := strings.TrimSpace(strings.ToLower(scanner.Text()))
			switch input {
			case "c", "continue", "":
				return // Continuar
			case "q", "quit":
				fmt.Println("👋 ¡Hasta luego! Buen trabajo.")
				os.Exit(0)
			default:
				fmt.Printf("❌ Escribe 'c' para continuar o 'q' para salir > ")
			}
		}
	}
}

func (s *Session) showHelp() {
	// No usar fmt.Scanln() que interrumpe nuestra goroutine
	fmt.Print("\r\033[K") // Limpiar línea actual
	fmt.Println()
	fmt.Println("+================================+")
	fmt.Println("|            AYUDA               |")
	fmt.Println("+================================+")
	fmt.Println()
	fmt.Println("🎮 Controles disponibles:")
	fmt.Println("   p, pause    - Pausar el timer actual")
	fmt.Println("   r, resume   - Reanudar el timer pausado")
	fmt.Println("   s, skip     - Saltar al siguiente período")
	fmt.Println("   q, quit     - Salir del programa")
	fmt.Println("   h, help     - Mostrar esta ayuda")
	fmt.Println()
	fmt.Println("💡 Consejos:")
	fmt.Println("   • Escribe el comando y presiona Enter")
	fmt.Println("   • El timer continúa corriendo mientras escribes")
	fmt.Println()
	fmt.Println("✅ Continúa escribiendo comandos...")
	fmt.Println()
}
