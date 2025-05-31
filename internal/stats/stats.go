package stats

import (
	"fmt"
	"time"
)

// SessionStats almacena las estadísticas de la sesión actual
type SessionStats struct {
	// Contadores básicos
	PomodorosCompleted  int // Pomodoros completados
	PomodorosSkipped    int // Pomodoros saltados
	BreaksCompleted     int // Descansos completados
	BreaksSkipped       int // Descansos saltados
	LongBreaksCompleted int // Descansos largos completados

	// Tiempo
	TotalWorkTime      time.Duration // Tiempo total trabajado
	TotalBreakTime     time.Duration // Tiempo total de descanso
	SessionStartTime   time.Time     // Cuando empezó la sesión
	CurrentStreakCount int           // Racha actual de pomodoros
	BestStreakCount    int           // Mejor racha de la sesión

	// Detalle de sesiones
	CompletedSessions []CompletedSession // Historial de sesiones completadas
}

// CompletedSession representa una sesión individual completada
type CompletedSession struct {
	Type       string        // "TRABAJO", "DESCANSO", "DESCANSO LARGO"
	Duration   time.Duration // Duración configurada
	ActualTime time.Duration // Tiempo real transcurrido
	StartTime  time.Time     // Cuando empezó
	EndTime    time.Time     // Cuando terminó
	Completed  bool          // true si se completó, false si se saltó
}

// NewSessionStats crea una nueva instancia de estadísticas
func NewSessionStats() *SessionStats {
	return &SessionStats{
		SessionStartTime:  time.Now(),
		CompletedSessions: make([]CompletedSession, 0),
	}
}

// AddCompletedPomodoro registra un pomodoro completado
func (s *SessionStats) AddCompletedPomodoro(duration, actualTime time.Duration, startTime, endTime time.Time) {
	s.PomodorosCompleted++
	s.TotalWorkTime += actualTime
	s.CurrentStreakCount++

	// Actualizar mejor racha
	if s.CurrentStreakCount > s.BestStreakCount {
		s.BestStreakCount = s.CurrentStreakCount
	}

	// Agregar al historial
	session := CompletedSession{
		Type:       "TRABAJO",
		Duration:   duration,
		ActualTime: actualTime,
		StartTime:  startTime,
		EndTime:    endTime,
		Completed:  true,
	}
	s.CompletedSessions = append(s.CompletedSessions, session)
}

// AddSkippedPomodoro registra un pomodoro saltado
func (s *SessionStats) AddSkippedPomodoro(duration, actualTime time.Duration, startTime, endTime time.Time) {
	s.PomodorosSkipped++
	s.TotalWorkTime += actualTime
	s.CurrentStreakCount = 0 // Rompe la racha

	// Agregar al historial
	session := CompletedSession{
		Type:       "TRABAJO",
		Duration:   duration,
		ActualTime: actualTime,
		StartTime:  startTime,
		EndTime:    endTime,
		Completed:  false,
	}
	s.CompletedSessions = append(s.CompletedSessions, session)
}

// AddCompletedBreak registra un descanso completado
func (s *SessionStats) AddCompletedBreak(breakType string, duration, actualTime time.Duration, startTime, endTime time.Time) {
	s.BreaksCompleted++
	s.TotalBreakTime += actualTime

	if breakType == "DESCANSO LARGO" {
		s.LongBreaksCompleted++
	}

	// Agregar al historial
	session := CompletedSession{
		Type:       breakType,
		Duration:   duration,
		ActualTime: actualTime,
		StartTime:  startTime,
		EndTime:    endTime,
		Completed:  true,
	}
	s.CompletedSessions = append(s.CompletedSessions, session)
}

// AddSkippedBreak registra un descanso saltado
func (s *SessionStats) AddSkippedBreak(breakType string, duration, actualTime time.Duration, startTime, endTime time.Time) {
	s.BreaksSkipped++
	s.TotalBreakTime += actualTime

	// Agregar al historial
	session := CompletedSession{
		Type:       breakType,
		Duration:   duration,
		ActualTime: actualTime,
		StartTime:  startTime,
		EndTime:    endTime,
		Completed:  false,
	}
	s.CompletedSessions = append(s.CompletedSessions, session)
}

// GetTotalSessions retorna el total de sesiones (completadas + saltadas)
func (s *SessionStats) GetTotalSessions() int {
	return s.PomodorosCompleted + s.PomodorosSkipped + s.BreaksCompleted + s.BreaksSkipped
}

// GetWorkEfficiency calcula el porcentaje de pomodoros completados vs saltados
func (s *SessionStats) GetWorkEfficiency() float64 {
	total := s.PomodorosCompleted + s.PomodorosSkipped
	if total == 0 {
		return 0
	}
	return float64(s.PomodorosCompleted) / float64(total) * 100
}

// GetSessionDuration retorna cuánto tiempo ha durado la sesión total
func (s *SessionStats) GetSessionDuration() time.Duration {
	return time.Since(s.SessionStartTime)
}

// FormatDuration convierte duración a formato legible
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// GetStatsDisplay retorna un string formateado con las estadísticas
func (s *SessionStats) GetStatsDisplay() string {
	efficiency := s.GetWorkEfficiency()
	sessionTime := s.GetSessionDuration()

	stats := fmt.Sprintf(`
+================================+
|          ESTADÍSTICAS          |
+================================+

📊 Resumen de la sesión:
   • Pomodoros completados: %d
   • Pomodoros saltados: %d
   • Descansos completados: %d
   • Descansos saltados: %d
   • Descansos largos: %d

🔥 Rachas:
   • Racha actual: %d pomodoros
   • Mejor racha: %d pomodoros

⏱️  Tiempo:
   • Tiempo trabajado: %s
   • Tiempo de descanso: %s
   • Duración de sesión: %s

📈 Eficiencia:
   • Eficiencia de trabajo: %.1f%%
   • Total de sesiones: %d

🎯 Productividad:
`,
		s.PomodorosCompleted,
		s.PomodorosSkipped,
		s.BreaksCompleted,
		s.BreaksSkipped,
		s.LongBreaksCompleted,
		s.CurrentStreakCount,
		s.BestStreakCount,
		FormatDuration(s.TotalWorkTime),
		FormatDuration(s.TotalBreakTime),
		FormatDuration(sessionTime),
		efficiency,
		s.GetTotalSessions())

	// Añadir barra de progreso visual para eficiencia
	if s.GetTotalSessions() > 0 {
		efficiencyBar := getEfficiencyBar(efficiency)
		stats += fmt.Sprintf("   • Progreso: [%s] %.1f%%\n", efficiencyBar, efficiency)
	}

	return stats
}

// getEfficiencyBar crea una barra visual de eficiencia
func getEfficiencyBar(efficiency float64) string {
	width := 20
	filled := int(efficiency / 100 * float64(width))

	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	var bar string
	for i := 0; i < width; i++ {
		if i < filled {
			if efficiency >= 80 {
				bar += "█" // Verde para alta eficiencia
			} else if efficiency >= 60 {
				bar += "▓" // Amarillo para eficiencia media
			} else {
				bar += "▒" // Rojo para baja eficiencia
			}
		} else {
			bar += "░"
		}
	}

	return bar
}

// GetQuickStats retorna estadísticas resumidas para mostrar durante el timer
func (s *SessionStats) GetQuickStats() string {
	return fmt.Sprintf("🍅 %d | 🔥 %d | ⏱️ %s",
		s.PomodorosCompleted,
		s.CurrentStreakCount,
		FormatDuration(s.TotalWorkTime))
}
