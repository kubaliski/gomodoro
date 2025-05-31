package session

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/config"
	"github.com/kubaliski/pomodoro-cli/internal/stats"
	"github.com/kubaliski/pomodoro-cli/internal/timer"
	"github.com/kubaliski/pomodoro-cli/internal/ui"
)

// Session maneja una sesión completa de pomodoros
type Session struct {
	Config          *config.Config
	PomodoroCount   int
	inputReader     *bufio.Reader
	globalInputChan chan string
	sessionStats    *stats.SessionStats
}

// NewSession crea una nueva sesión
func NewSession(cfg *config.Config) *Session {
	session := &Session{
		Config:          cfg,
		PomodoroCount:   0,
		inputReader:     bufio.NewReader(os.Stdin),
		globalInputChan: make(chan string, 10),
		sessionStats:    stats.NewSessionStats(),
	}

	// UNA SOLA goroutine global para todo el input
	go session.startGlobalInputListener()

	return session
}

// startGlobalInputListener es la ÚNICA goroutine que lee input
func (s *Session) startGlobalInputListener() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if input != "" {
			select {
			case s.globalInputChan <- input:
			default:
				// Canal lleno, ignorar
			}
		}
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
			if !s.showWorkCompleted(breakType, breakDuration) {
				return // Usuario salió
			}
		} else if result == TimerResultSkipped {
			if !s.showWorkSkipped(breakType, breakDuration) {
				return // Usuario salió
			}
		}

		// Reiniciar display
		ui.ResetDisplay()

		// Sesión de descanso
		result = s.runTimerWithControls(breakDuration, breakType)
		if result == TimerResultQuit {
			fmt.Println("\n👋 ¡Hasta luego! Buen trabajo.")
			return
		}

		// Mostrar mensaje de descanso completado
		if result == TimerResultCompleted {
			if !s.showBreakCompleted(breakType) {
				return // Usuario salió
			}
		} else if result == TimerResultSkipped {
			if !s.showBreakSkipped(breakType) {
				return // Usuario salió
			}
		}

		// Reiniciar display
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

	// Trackear tiempo de inicio para estadísticas
	sessionStartTime := time.Now()

	// Ticker para actualizar cada segundo
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Display inicial con estadísticas rápidas
	fmt.Print("\r\033[K")
	s.displayTimerWithStats(pomodoroTimer.Remaining, state, pomodoroTimer.GetStatus(), duration)
	fmt.Println()
	fmt.Print("Comando > ")

	for pomodoroTimer.IsRunning && !pomodoroTimer.IsFinished() {
		select {
		case <-ticker.C:
			// Actualizar timer
			pomodoroTimer.Tick()

			// Guardar posición actual del cursor
			fmt.Print("\033[s") // Guardar cursor

			// Ir arriba y actualizar timer (sin tocar la línea de comando)
			fmt.Print("\033[A")   // Subir una línea
			fmt.Print("\r\033[K") // Limpiar línea del timer
			s.displayTimerWithStats(pomodoroTimer.Remaining, state, pomodoroTimer.GetStatus(), duration)

			// Restaurar posición del cursor (línea de comandos)
			fmt.Print("\033[u") // Restaurar cursor

		case input := <-s.globalInputChan:
			// Mostrar el comando que se escribió
			fmt.Printf("%s\n", input)

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

				// Registrar en estadísticas
				sessionEndTime := time.Now()
				actualTime := sessionEndTime.Sub(sessionStartTime)
				if state == "TRABAJO" {
					s.sessionStats.AddSkippedPomodoro(duration, actualTime, sessionStartTime, sessionEndTime)
				} else {
					s.sessionStats.AddSkippedBreak(state, duration, actualTime, sessionStartTime, sessionEndTime)
				}

				return TimerResultSkipped

			case "q", "quit":
				fmt.Println("👋 Saliendo...")
				return TimerResultQuit

			case "h", "help":
				s.showInlineHelp()

			case "stats", "estadisticas":
				s.showDetailedStats()

			default:
				fmt.Printf("❌ Comando '%s' no reconocido. Usa: p, r, s, q, h, stats\n", input)
			}

			// Nuevo prompt para el siguiente comando
			fmt.Print("Comando > ")

		default:
			// No bloquear
		}

		// Verificar si fue saltado o se quiere salir
		if pomodoroTimer.IsSkipped {
			return TimerResultSkipped
		}
	}

	// Timer completado - registrar en estadísticas
	sessionEndTime := time.Now()
	actualTime := sessionEndTime.Sub(sessionStartTime)

	if state == "TRABAJO" {
		s.sessionStats.AddCompletedPomodoro(duration, actualTime, sessionStartTime, sessionEndTime)
	} else {
		s.sessionStats.AddCompletedBreak(state, duration, actualTime, sessionStartTime, sessionEndTime)
	}

	fmt.Println() // Nueva línea al terminar
	return TimerResultCompleted
}

func (s *Session) showInlineHelp() {
	fmt.Println()
	fmt.Println("🎮 Controles: (p)ausar (r)eanudar (s)altar (q)salir (stats)estadísticas")
	fmt.Println("💡 Escribe el comando y presiona Enter")
	fmt.Println()
}

// displayTimerWithStats muestra el timer junto con estadísticas rápidas
func (s *Session) displayTimerWithStats(remaining time.Duration, state string, status string, duration time.Duration) {
	// Timer principal
	ui.DisplayTimer(remaining, state, status, duration)

	// Estadísticas rápidas en la misma línea
	quickStats := s.sessionStats.GetQuickStats()
	fmt.Printf(" | %s", quickStats)
}

// showDetailedStats muestra estadísticas detalladas
func (s *Session) showDetailedStats() {
	fmt.Println()
	fmt.Print(s.sessionStats.GetStatsDisplay())
	fmt.Println("\n⏸️  Presiona Enter para continuar...")

	// Esperar que el usuario presione Enter usando nuestro sistema global
	for {
		select {
		case input := <-s.globalInputChan:
			if input == "" || input == "c" || input == "continue" {
				return
			}
			fmt.Println("⏸️  Presiona Enter para continuar...")
		default:
			time.Sleep(10 * time.Millisecond)
		}
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
	fmt.Println("🎮 Controles: (p)ausar (r)eanudar (s)altar (q)salir (h)ayuda")
	fmt.Println("📊 Nuevo: (stats) para ver estadísticas detalladas")
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

func (s *Session) showWorkCompleted(nextBreakType string, breakDuration time.Duration) bool {
	fmt.Println()
	fmt.Println("+================================+")
	fmt.Println("|       POMODORO COMPLETO!       |")
	fmt.Println("+================================+")
	fmt.Printf("✅ ¡Pomodoro #%d completado!\n", s.PomodoroCount)
	fmt.Printf("🎯 Próximo: %s (%s)\n", nextBreakType, ui.FormatDuration(breakDuration))
	fmt.Println()

	// Mostrar estadísticas rápidas
	fmt.Printf("📊 Estadísticas: %s\n", s.sessionStats.GetQuickStats())
	if s.sessionStats.CurrentStreakCount > 1 {
		fmt.Printf("🔥 ¡Racha de %d pomodoros!\n", s.sessionStats.CurrentStreakCount)
	}
	fmt.Println()

	fmt.Println("Escribe 'c' para continuar, 'stats' para ver estadísticas detalladas, o 'q' para salir")

	return s.waitForInputWithStats([]string{"c", "continue"}, []string{"q", "quit"})
}

func (s *Session) showWorkSkipped(nextBreakType string, breakDuration time.Duration) bool {
	fmt.Println()
	fmt.Println("+================================+")
	fmt.Println("|      POMODORO SALTADO!         |")
	fmt.Println("+================================+")
	fmt.Printf("⏭️  Pomodoro #%d saltado\n", s.PomodoroCount)
	fmt.Printf("🎯 Próximo: %s (%s)\n", nextBreakType, ui.FormatDuration(breakDuration))
	fmt.Println()
	fmt.Println("Escribe 'c' para continuar o 'q' para salir")

	return s.waitForInput([]string{"c", "continue"}, []string{"q", "quit"})
}

func (s *Session) showBreakCompleted(breakType string) bool {
	fmt.Println()
	fmt.Println("+================================+")
	fmt.Println("|      DESCANSO COMPLETADO!      |")
	fmt.Println("+================================+")
	fmt.Printf("✅ %s terminado\n", breakType)
	fmt.Println("💪 ¡Listo para el siguiente pomodoro!")
	fmt.Println()
	fmt.Println("Escribe 'c' para continuar o 'q' para salir")

	return s.waitForInput([]string{"c", "continue"}, []string{"q", "quit"})
}

func (s *Session) showBreakSkipped(breakType string) bool {
	fmt.Println()
	fmt.Println("+================================+")
	fmt.Println("|      DESCANSO SALTADO!         |")
	fmt.Println("+================================+")
	fmt.Printf("⏭️  %s saltado\n", breakType)
	fmt.Println("💪 ¡Listo para el siguiente pomodoro!")
	fmt.Println()
	fmt.Println("Escribe 'c' para continuar o 'q' para salir")

	return s.waitForInput([]string{"c", "continue"}, []string{"q", "quit"})
}

// waitForInputWithStats es como waitForInput pero también acepta "stats"
func (s *Session) waitForInputWithStats(continueCommands, quitCommands []string) bool {
	fmt.Print("Comando > ")

	for {
		select {
		case input := <-s.globalInputChan:
			// Mostrar el comando escrito
			fmt.Printf("%s\n", input)

			// Verificar comando de estadísticas
			if input == "stats" || input == "estadisticas" {
				s.showDetailedStats()
				fmt.Println("Escribe 'c' para continuar, 'stats' para ver estadísticas detalladas, o 'q' para salir")
				fmt.Print("Comando > ")
				continue
			}

			// Verificar comandos de continuar
			for _, cmd := range continueCommands {
				if input == cmd || input == "" {
					return true // Continuar
				}
			}

			// Verificar comandos de salir
			for _, cmd := range quitCommands {
				if input == cmd {
					fmt.Println("👋 ¡Hasta luego! Buen trabajo.")
					return false // Salir
				}
			}

			// Comando no reconocido
			fmt.Printf("❌ Escribe 'c' para continuar, 'stats' para estadísticas, o 'q' para salir\n")
			fmt.Print("Comando > ")

		default:
			// No bloquear
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// waitForInput espera comandos específicos usando el canal global (versión básica)
func (s *Session) waitForInput(continueCommands, quitCommands []string) bool {
	fmt.Print("Comando > ")

	for {
		select {
		case input := <-s.globalInputChan:
			// Mostrar el comando escrito
			fmt.Printf("%s\n", input)

			// Verificar comandos de continuar
			for _, cmd := range continueCommands {
				if input == cmd || input == "" {
					return true // Continuar
				}
			}

			// Verificar comandos de salir
			for _, cmd := range quitCommands {
				if input == cmd {
					fmt.Println("👋 ¡Hasta luego! Buen trabajo.")
					return false // Salir
				}
			}

			// Comando no reconocido
			fmt.Printf("❌ Escribe 'c' para continuar o 'q' para salir\n")
			fmt.Print("Comando > ")

		default:
			// No bloquear
			time.Sleep(10 * time.Millisecond)
		}
	}
}
