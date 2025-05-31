package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/kubaliski/pomodoro-core/stats"
)

// StatsDisplayConfig configura el display de estadísticas
type StatsDisplayConfig struct {
	ShowGraphs  bool
	ShowTrends  bool
	ShowDetails bool
	GraphWidth  int
	UseColors   bool
	CompactMode bool
}

// DefaultStatsConfig retorna configuración por defecto
func DefaultStatsConfig() StatsDisplayConfig {
	return StatsDisplayConfig{
		ShowGraphs:  true,
		ShowTrends:  true,
		ShowDetails: true,
		GraphWidth:  40,
		UseColors:   true,
		CompactMode: false,
	}
}

// EnhancedStatsDisplay genera un display avanzado de estadísticas
func EnhancedStatsDisplay(statsManager *stats.SessionStats, config StatsDisplayConfig) string {
	snapshot := statsManager.GetSnapshot()
	var result strings.Builder

	if config.CompactMode {
		return compactStatsDisplay(snapshot)
	}

	// Header principal
	result.WriteString(Colorize("╔══════════════════════════════════════════════════════════════╗\n", ColorCyan, config.UseColors))
	result.WriteString(Colorize("║                    📊 ESTADÍSTICAS POMODORO                  ║\n", ColorCyan, config.UseColors))
	result.WriteString(Colorize("╚══════════════════════════════════════════════════════════════╝\n", ColorCyan, config.UseColors))
	result.WriteString("\n")

	// Sección de resumen principal
	result.WriteString(buildSummarySection(snapshot, config))
	result.WriteString("\n")

	// Sección de rendimiento
	if config.ShowDetails {
		result.WriteString(buildPerformanceSection(snapshot, config))
		result.WriteString("\n")
	}

	// Gráficos de productividad
	if config.ShowGraphs {
		result.WriteString(buildProductivityGraphs(snapshot, config))
		result.WriteString("\n")
	}

	// Tendencias y análisis
	if config.ShowTrends {
		result.WriteString(buildTrendsSection(snapshot, config))
		result.WriteString("\n")
	}

	// Footer con consejos
	result.WriteString(buildTipsSection(snapshot, config))

	return result.String()
}

// buildSummarySection construye la sección de resumen
func buildSummarySection(snapshot stats.StatsSnapshot, config StatsDisplayConfig) string {
	var result strings.Builder

	result.WriteString(Colorize("🍅 RESUMEN DE SESIÓN\n", ColorYellow, config.UseColors))
	result.WriteString(Colorize("─────────────────────\n", ColorGray, config.UseColors))

	// Métricas principales en columnas
	col1 := fmt.Sprintf("🍅 Pomodoros: %s%d%s",
		ColorStart(ColorGreen, config.UseColors),
		snapshot.PomodorosCompleted,
		ColorEnd(config.UseColors))

	col2 := fmt.Sprintf("⏭️  Saltados: %s%d%s",
		ColorStart(ColorRed, config.UseColors),
		snapshot.PomodorosSkipped,
		ColorEnd(config.UseColors))

	result.WriteString(fmt.Sprintf("%-30s %s\n", col1, col2))

	col3 := fmt.Sprintf("🔥 Racha actual: %s%d%s",
		ColorStart(ColorOrange, config.UseColors),
		snapshot.CurrentStreak,
		ColorEnd(config.UseColors))

	col4 := fmt.Sprintf("🏆 Mejor racha: %s%d%s",
		ColorStart(ColorPurple, config.UseColors),
		snapshot.BestStreak,
		ColorEnd(config.UseColors))

	result.WriteString(fmt.Sprintf("%-30s %s\n", col3, col4))

	// Tiempo total con formato amigable
	workTime := formatDurationDetailed(snapshot.TotalWorkTime)
	breakTime := formatDurationDetailed(snapshot.TotalBreakTime)
	sessionTime := formatDurationDetailed(snapshot.SessionDuration)

	result.WriteString(fmt.Sprintf("⏱️  Tiempo trabajo: %s%s%s\n",
		ColorStart(ColorBlue, config.UseColors), workTime, ColorEnd(config.UseColors)))
	result.WriteString(fmt.Sprintf("🧘 Tiempo descanso: %s%s%s\n",
		ColorStart(ColorCyan, config.UseColors), breakTime, ColorEnd(config.UseColors)))
	result.WriteString(fmt.Sprintf("📅 Duración sesión: %s%s%s\n",
		ColorStart(ColorMagenta, config.UseColors), sessionTime, ColorEnd(config.UseColors)))

	return result.String()
}

// buildPerformanceSection construye la sección de rendimiento
func buildPerformanceSection(snapshot stats.StatsSnapshot, config StatsDisplayConfig) string {
	var result strings.Builder

	result.WriteString(Colorize("📈 ANÁLISIS DE RENDIMIENTO\n", ColorYellow, config.UseColors))
	result.WriteString(Colorize("─────────────────────────\n", ColorGray, config.UseColors))

	// Eficiencia de trabajo
	efficiency := snapshot.WorkEfficiency
	efficiencyBar := createProgressBar(efficiency/100.0, 20, config.UseColors)
	result.WriteString(fmt.Sprintf("💪 Eficiencia trabajo: %s %.1f%%\n", efficiencyBar, efficiency))

	// Ratio descansos
	totalBreaks := snapshot.BreaksCompleted + snapshot.BreaksSkipped
	var breakEfficiency float64
	if totalBreaks > 0 {
		breakEfficiency = (float64(snapshot.BreaksCompleted) / float64(totalBreaks)) * 100
	}
	breakBar := createProgressBar(breakEfficiency/100.0, 20, config.UseColors)
	result.WriteString(fmt.Sprintf("🧘 Descansos tomados: %s %.1f%%\n", breakBar, breakEfficiency))

	// Tiempo promedio por pomodoro
	if snapshot.PomodorosCompleted > 0 {
		avgPomodoroTime := snapshot.TotalWorkTime / time.Duration(snapshot.PomodorosCompleted)
		result.WriteString(fmt.Sprintf("⏱️  Tiempo promedio/pomodoro: %s\n", formatDurationDetailed(avgPomodoroTime)))
	}

	// Velocidad de la sesión
	if snapshot.SessionDuration > 0 {
		pomodorosPerHour := float64(snapshot.PomodorosCompleted) / snapshot.SessionDuration.Hours()
		result.WriteString(fmt.Sprintf("🚀 Velocidad: %.1f pomodoros/hora\n", pomodorosPerHour))
	}

	return result.String()
}

// buildProductivityGraphs construye gráficos de productividad
func buildProductivityGraphs(snapshot stats.StatsSnapshot, config StatsDisplayConfig) string {
	var result strings.Builder

	result.WriteString(Colorize("📊 GRÁFICO DE PRODUCTIVIDAD\n", ColorYellow, config.UseColors))
	result.WriteString(Colorize("─────────────────────────\n", ColorGray, config.UseColors))

	// Gráfico de distribución de tiempo
	totalTime := snapshot.TotalWorkTime + snapshot.TotalBreakTime
	if totalTime > 0 {
		workRatio := float64(snapshot.TotalWorkTime) / float64(totalTime)
		breakRatio := float64(snapshot.TotalBreakTime) / float64(totalTime)

		workChars := int(workRatio * float64(config.GraphWidth))
		breakChars := config.GraphWidth - workChars

		result.WriteString("Distribución de tiempo:\n")
		result.WriteString(Colorize("Trabajo  ", ColorBlue, config.UseColors))
		result.WriteString(Colorize(strings.Repeat("█", workChars), ColorBlue, config.UseColors))
		result.WriteString(Colorize(strings.Repeat("░", breakChars), ColorCyan, config.UseColors))
		result.WriteString(fmt.Sprintf(" %.1f%%\n", workRatio*100))

		result.WriteString(Colorize("Descanso ", ColorCyan, config.UseColors))
		result.WriteString(Colorize(strings.Repeat("░", workChars), ColorBlue, config.UseColors))
		result.WriteString(Colorize(strings.Repeat("█", breakChars), ColorCyan, config.UseColors))
		result.WriteString(fmt.Sprintf(" %.1f%%\n", breakRatio*100))
	}

	// Gráfico de racha visual
	if snapshot.CurrentStreak > 0 {
		result.WriteString("\nRacha actual:\n")
		streakDisplay := createStreakVisualization(snapshot.CurrentStreak, snapshot.BestStreak, config)
		result.WriteString(streakDisplay)
	}

	return result.String()
}

// buildTrendsSection construye la sección de tendencias
func buildTrendsSection(snapshot stats.StatsSnapshot, config StatsDisplayConfig) string {
	var result strings.Builder

	result.WriteString(Colorize("📈 TENDENCIAS Y LOGROS\n", ColorYellow, config.UseColors))
	result.WriteString(Colorize("────────────────────\n", ColorGray, config.UseColors))

	// Logros desbloqueados
	achievements := calculateAchievements(snapshot)
	if len(achievements) > 0 {
		result.WriteString("🏆 Logros desbloqueados:\n")
		for _, achievement := range achievements {
			result.WriteString(fmt.Sprintf("   %s %s\n", achievement.Icon, achievement.Description))
		}
	}

	// Proyecciones
	if snapshot.PomodorosCompleted > 0 && snapshot.SessionDuration.Hours() > 0 {
		result.WriteString("\n🔮 Proyecciones:\n")

		hourlyRate := float64(snapshot.PomodorosCompleted) / snapshot.SessionDuration.Hours()
		if hourlyRate > 0 {
			pomodorosIn8h := int(hourlyRate * 8)
			result.WriteString(fmt.Sprintf("   En 8 horas: ~%d pomodoros\n", pomodorosIn8h))

			timeFor25 := time.Duration(float64(25*time.Hour) / hourlyRate)
			result.WriteString(fmt.Sprintf("   Para 25 pomodoros: ~%s\n", formatDurationDetailed(timeFor25)))
		}
	}

	return result.String()
}

// buildTipsSection construye la sección de consejos
func buildTipsSection(snapshot stats.StatsSnapshot, config StatsDisplayConfig) string {
	var result strings.Builder

	result.WriteString(Colorize("💡 CONSEJOS PERSONALIZADOS\n", ColorYellow, config.UseColors))
	result.WriteString(Colorize("─────────────────────────\n", ColorGray, config.UseColors))

	tips := generatePersonalizedTips(snapshot)
	for _, tip := range tips {
		result.WriteString(fmt.Sprintf("   %s\n", tip))
	}

	return result.String()
}

// Helper functions

func compactStatsDisplay(snapshot stats.StatsSnapshot) string {
	return fmt.Sprintf("🍅 %d | 🔥 %d | ⏱️ %s | 📈 %.1f%%",
		snapshot.PomodorosCompleted,
		snapshot.CurrentStreak,
		formatDurationDetailed(snapshot.TotalWorkTime),
		snapshot.WorkEfficiency)
}

func createProgressBar(progress float64, width int, useColors bool) string {
	filled := int(progress * float64(width))
	empty := width - filled

	bar := Colorize(strings.Repeat("█", filled), ColorGreen, useColors) +
		Colorize(strings.Repeat("░", empty), ColorGray, useColors)

	return fmt.Sprintf("[%s]", bar)
}

func createStreakVisualization(current, best int, config StatsDisplayConfig) string {
	maxDisplay := 20
	streakChars := int(math.Min(float64(current), float64(maxDisplay)))

	streak := Colorize(strings.Repeat("🔥", streakChars), ColorOrange, config.UseColors)
	if current > maxDisplay {
		streak += Colorize(fmt.Sprintf(" +%d", current-maxDisplay), ColorRed, config.UseColors)
	}

	progress := ""
	if best > 0 {
		progressRatio := float64(current) / float64(best)
		progress = fmt.Sprintf(" (%.1f%% del mejor)", progressRatio*100)
	}

	return streak + progress + "\n"
}

// Achievement representa un logro
type Achievement struct {
	Icon        string
	Description string
}

func calculateAchievements(snapshot stats.StatsSnapshot) []Achievement {
	var achievements []Achievement

	if snapshot.PomodorosCompleted >= 1 {
		achievements = append(achievements, Achievement{"🌱", "Primer pomodoro completado"})
	}
	if snapshot.PomodorosCompleted >= 5 {
		achievements = append(achievements, Achievement{"🌿", "5 pomodoros - Construyendo hábito"})
	}
	if snapshot.PomodorosCompleted >= 25 {
		achievements = append(achievements, Achievement{"🌳", "25 pomodoros - Árbol de productividad"})
	}
	if snapshot.CurrentStreak >= 3 {
		achievements = append(achievements, Achievement{"🔥", "Racha de fuego - 3 consecutivos"})
	}
	if snapshot.CurrentStreak >= 10 {
		achievements = append(achievements, Achievement{"💥", "Racha explosiva - 10 consecutivos"})
	}
	if snapshot.WorkEfficiency >= 90 {
		achievements = append(achievements, Achievement{"⚡", "Máxima eficiencia - 90%+"})
	}
	if snapshot.TotalWorkTime >= 2*time.Hour {
		achievements = append(achievements, Achievement{"⏰", "Maratonista - 2+ horas de trabajo"})
	}

	return achievements
}

func generatePersonalizedTips(snapshot stats.StatsSnapshot) []string {
	var tips []string

	if snapshot.WorkEfficiency < 70 {
		tips = append(tips, "💪 Intenta minimizar interrupciones para mejorar tu eficiencia")
	}

	if snapshot.BreaksSkipped > snapshot.BreaksCompleted {
		tips = append(tips, "🧘 Los descansos son importantes - mejoran tu productividad")
	}

	if snapshot.CurrentStreak == 0 && snapshot.PomodorosCompleted > 0 {
		tips = append(tips, "🎯 Mantén el ritmo - cada pomodoro cuenta para tu racha")
	}

	if snapshot.CurrentStreak >= 5 {
		tips = append(tips, "🔥 ¡Excelente racha! Mantén este impulso")
	}

	if snapshot.TotalWorkTime > 3*time.Hour {
		tips = append(tips, "🎉 Sesión extensa - considera tomar un descanso más largo")
	}

	if len(tips) == 0 {
		tips = append(tips, "✨ ¡Sigue así! Cada pomodoro te acerca a tus objetivos")
	}

	return tips
}

func formatDurationDetailed(d time.Duration) string {
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
