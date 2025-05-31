package stats

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// SessionStats maneja las estadísticas de la sesión de forma thread-safe
type SessionStats struct {
	mu sync.RWMutex

	// Contadores básicos
	PomodorosCompleted  int `json:"pomodoros_completed"`
	PomodorosSkipped    int `json:"pomodoros_skipped"`
	BreaksCompleted     int `json:"breaks_completed"`
	BreaksSkipped       int `json:"breaks_skipped"`
	LongBreaksCompleted int `json:"long_breaks_completed"`

	// Tiempo
	TotalWorkTime      time.Duration `json:"total_work_time"`
	TotalBreakTime     time.Duration `json:"total_break_time"`
	SessionStartTime   time.Time     `json:"session_start_time"`
	CurrentStreakCount int           `json:"current_streak_count"`
	BestStreakCount    int           `json:"best_streak_count"`

	// Historial de sesiones
	CompletedSessions []CompletedSession `json:"completed_sessions"`
}

// CompletedSession representa una sesión individual completada
type CompletedSession struct {
	Type       string        `json:"type"`        // "TRABAJO", "DESCANSO", "DESCANSO LARGO"
	Duration   time.Duration `json:"duration"`    // Duración configurada
	ActualTime time.Duration `json:"actual_time"` // Tiempo real transcurrido
	StartTime  time.Time     `json:"start_time"`  // Cuando empezó
	EndTime    time.Time     `json:"end_time"`    // Cuando terminó
	Completed  bool          `json:"completed"`   // true si se completó, false si se saltó
}

// StatsSnapshot representa una instantánea inmutable de las estadísticas
type StatsSnapshot struct {
	PomodorosCompleted  int
	PomodorosSkipped    int
	BreaksCompleted     int
	BreaksSkipped       int
	LongBreaksCompleted int
	CurrentStreak       int
	BestStreak          int
	TotalWorkTime       time.Duration
	TotalBreakTime      time.Duration
	SessionDuration     time.Duration
	WorkEfficiency      float64
	TotalSessions       int
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
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.mu.Lock()
	defer s.mu.Unlock()

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

// GetSnapshot retorna una instantánea inmutable de las estadísticas actuales
func (s *SessionStats) GetSnapshot() StatsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return StatsSnapshot{
		PomodorosCompleted:  s.PomodorosCompleted,
		PomodorosSkipped:    s.PomodorosSkipped,
		BreaksCompleted:     s.BreaksCompleted,
		BreaksSkipped:       s.BreaksSkipped,
		LongBreaksCompleted: s.LongBreaksCompleted,
		CurrentStreak:       s.CurrentStreakCount,
		BestStreak:          s.BestStreakCount,
		TotalWorkTime:       s.TotalWorkTime,
		TotalBreakTime:      s.TotalBreakTime,
		SessionDuration:     time.Since(s.SessionStartTime),
		WorkEfficiency:      s.calculateWorkEfficiency(),
		TotalSessions:       s.getTotalSessions(),
	}
}

// GetTotalSessions retorna el total de sesiones (debe llamarse con lock)
func (s *SessionStats) getTotalSessions() int {
	return s.PomodorosCompleted + s.PomodorosSkipped + s.BreaksCompleted + s.BreaksSkipped
}

// GetWorkEfficiency calcula el porcentaje de pomodoros completados vs saltados (debe llamarse con lock)
func (s *SessionStats) calculateWorkEfficiency() float64 {
	total := s.PomodorosCompleted + s.PomodorosSkipped
	if total == 0 {
		return 0
	}
	return float64(s.PomodorosCompleted) / float64(total) * 100
}

// GetSessionDuration retorna cuánto tiempo ha durado la sesión total
func (s *SessionStats) GetSessionDuration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.SessionStartTime)
}

// Reset reinicia todas las estadísticas
func (s *SessionStats) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.PomodorosCompleted = 0
	s.PomodorosSkipped = 0
	s.BreaksCompleted = 0
	s.BreaksSkipped = 0
	s.LongBreaksCompleted = 0
	s.TotalWorkTime = 0
	s.TotalBreakTime = 0
	s.SessionStartTime = time.Now()
	s.CurrentStreakCount = 0
	s.BestStreakCount = 0
	s.CompletedSessions = make([]CompletedSession, 0)
}

// ExportJSON exporta las estadísticas completas a JSON
func (s *SessionStats) ExportJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Crear estructura sin mutex para exportar
	data := struct {
		PomodorosCompleted  int                `json:"pomodoros_completed"`
		PomodorosSkipped    int                `json:"pomodoros_skipped"`
		BreaksCompleted     int                `json:"breaks_completed"`
		BreaksSkipped       int                `json:"breaks_skipped"`
		LongBreaksCompleted int                `json:"long_breaks_completed"`
		TotalWorkTime       time.Duration      `json:"total_work_time"`
		TotalBreakTime      time.Duration      `json:"total_break_time"`
		SessionStartTime    time.Time          `json:"session_start_time"`
		CurrentStreakCount  int                `json:"current_streak_count"`
		BestStreakCount     int                `json:"best_streak_count"`
		CompletedSessions   []CompletedSession `json:"completed_sessions"`
		ExportedAt          time.Time          `json:"exported_at"`
	}{
		PomodorosCompleted:  s.PomodorosCompleted,
		PomodorosSkipped:    s.PomodorosSkipped,
		BreaksCompleted:     s.BreaksCompleted,
		BreaksSkipped:       s.BreaksSkipped,
		LongBreaksCompleted: s.LongBreaksCompleted,
		TotalWorkTime:       s.TotalWorkTime,
		TotalBreakTime:      s.TotalBreakTime,
		SessionStartTime:    s.SessionStartTime,
		CurrentStreakCount:  s.CurrentStreakCount,
		BestStreakCount:     s.BestStreakCount,
		CompletedSessions:   s.CompletedSessions,
		ExportedAt:          time.Now(),
	}

	return json.MarshalIndent(data, "", "  ")
}

// ImportJSON importa estadísticas desde JSON
func (s *SessionStats) ImportJSON(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var imported struct {
		PomodorosCompleted  int                `json:"pomodoros_completed"`
		PomodorosSkipped    int                `json:"pomodoros_skipped"`
		BreaksCompleted     int                `json:"breaks_completed"`
		BreaksSkipped       int                `json:"breaks_skipped"`
		LongBreaksCompleted int                `json:"long_breaks_completed"`
		TotalWorkTime       time.Duration      `json:"total_work_time"`
		TotalBreakTime      time.Duration      `json:"total_break_time"`
		SessionStartTime    time.Time          `json:"session_start_time"`
		CurrentStreakCount  int                `json:"current_streak_count"`
		BestStreakCount     int                `json:"best_streak_count"`
		CompletedSessions   []CompletedSession `json:"completed_sessions"`
		ExportedAt          time.Time          `json:"exported_at"`
	}

	if err := json.Unmarshal(data, &imported); err != nil {
		return fmt.Errorf("failed to unmarshal stats: %w", err)
	}

	// Copiar datos importados campo por campo (evitar copiar mutex)
	s.PomodorosCompleted = imported.PomodorosCompleted
	s.PomodorosSkipped = imported.PomodorosSkipped
	s.BreaksCompleted = imported.BreaksCompleted
	s.BreaksSkipped = imported.BreaksSkipped
	s.LongBreaksCompleted = imported.LongBreaksCompleted
	s.TotalWorkTime = imported.TotalWorkTime
	s.TotalBreakTime = imported.TotalBreakTime
	s.SessionStartTime = imported.SessionStartTime
	s.CurrentStreakCount = imported.CurrentStreakCount
	s.BestStreakCount = imported.BestStreakCount
	s.CompletedSessions = imported.CompletedSessions

	return nil
}

// GetQuickStats retorna estadísticas resumidas para mostrar durante el timer
func (s *SessionStats) GetQuickStats() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return fmt.Sprintf("🍅 %d | 🔥 %d | ⏱️ %s",
		s.PomodorosCompleted,
		s.CurrentStreakCount,
		FormatDuration(s.TotalWorkTime))
}

// GetStatsDisplay retorna un string formateado con las estadísticas completas
func (s *SessionStats) GetStatsDisplay() string {
	snapshot := s.GetSnapshot()

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
		snapshot.PomodorosCompleted,
		snapshot.PomodorosSkipped,
		snapshot.BreaksCompleted,
		snapshot.BreaksSkipped,
		snapshot.LongBreaksCompleted,
		snapshot.CurrentStreak,
		snapshot.BestStreak,
		FormatDuration(snapshot.TotalWorkTime),
		FormatDuration(snapshot.TotalBreakTime),
		FormatDuration(snapshot.SessionDuration),
		snapshot.WorkEfficiency,
		snapshot.TotalSessions)

	// Añadir barra de progreso visual para eficiencia
	if snapshot.TotalSessions > 0 {
		efficiencyBar := getEfficiencyBar(snapshot.WorkEfficiency)
		stats += fmt.Sprintf("   • Progreso: [%s] %.1f%%\n", efficiencyBar, snapshot.WorkEfficiency)
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

// GetCompletedSessions retorna una copia de las sesiones completadas
func (s *SessionStats) GetCompletedSessions() []CompletedSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Crear copia para evitar modificaciones externas
	sessions := make([]CompletedSession, len(s.CompletedSessions))
	copy(sessions, s.CompletedSessions)
	return sessions
}

// GetRecentSessions retorna las últimas N sesiones
func (s *SessionStats) GetRecentSessions(count int) []CompletedSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if count <= 0 {
		return []CompletedSession{}
	}

	total := len(s.CompletedSessions)
	if count > total {
		count = total
	}

	start := total - count
	sessions := make([]CompletedSession, count)
	copy(sessions, s.CompletedSessions[start:])
	return sessions
}

// GetWorkSessions retorna solo las sesiones de trabajo (pomodoros)
func (s *SessionStats) GetWorkSessions() []CompletedSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var workSessions []CompletedSession
	for _, session := range s.CompletedSessions {
		if session.Type == "TRABAJO" {
			workSessions = append(workSessions, session)
		}
	}
	return workSessions
}

// GetBreakSessions retorna solo las sesiones de descanso
func (s *SessionStats) GetBreakSessions() []CompletedSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var breakSessions []CompletedSession
	for _, session := range s.CompletedSessions {
		if session.Type != "TRABAJO" {
			breakSessions = append(breakSessions, session)
		}
	}
	return breakSessions
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
