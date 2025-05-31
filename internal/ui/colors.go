package ui

import "strings"

// Color representa códigos de color ANSI
type Color string

// Definición de colores ANSI - VERSIÓN ÚNICA Y CONSOLIDADA
const (
	ColorReset Color = "\033[0m"
	ColorBold  Color = "\033[1m"
	ColorDim   Color = "\033[2m"

	// Colores de texto básicos
	ColorBlack   Color = "\033[30m"
	ColorRed     Color = "\033[31m"
	ColorGreen   Color = "\033[32m"
	ColorYellow  Color = "\033[33m"
	ColorBlue    Color = "\033[34m"
	ColorMagenta Color = "\033[35m"
	ColorCyan    Color = "\033[36m"
	ColorWhite   Color = "\033[37m"
	ColorGray    Color = "\033[90m"

	// Colores brillantes
	ColorBrightRed     Color = "\033[91m"
	ColorBrightGreen   Color = "\033[92m"
	ColorBrightYellow  Color = "\033[93m"
	ColorBrightBlue    Color = "\033[94m"
	ColorBrightMagenta Color = "\033[95m"
	ColorBrightCyan    Color = "\033[96m"
	ColorBrightWhite   Color = "\033[97m"

	// Colores personalizados para el tema pomodoro
	ColorOrange Color = "\033[38;5;208m" // Naranja para fuego/racha
	ColorPurple Color = "\033[38;5;135m" // Púrpura para logros
	ColorPink   Color = "\033[38;5;213m" // Rosa para destacados

	// Colores de fondo
	ColorBgRed     Color = "\033[41m"
	ColorBgGreen   Color = "\033[42m"
	ColorBgYellow  Color = "\033[43m"
	ColorBgBlue    Color = "\033[44m"
	ColorBgMagenta Color = "\033[45m"
	ColorBgCyan    Color = "\033[46m"
)

// Aliases para compatibilidad con display.go existente
const (
	BgRed    = ColorBgRed
	BgGreen  = ColorBgGreen
	BgYellow = ColorBgYellow
	BgBlue   = ColorBgBlue
	BgPurple = ColorBgMagenta
	BgCyan   = ColorBgCyan
)

// Colorize aplica un color al texto si useColors es true
func Colorize(text string, color Color, useColors bool) string {
	if !useColors {
		return text
	}
	return string(color) + text + string(ColorReset)
}

// ColorStart retorna el código de inicio de color
func ColorStart(color Color, useColors bool) string {
	if !useColors {
		return ""
	}
	return string(color)
}

// ColorEnd retorna el código de reset de color
func ColorEnd(useColors bool) string {
	if !useColors {
		return ""
	}
	return string(ColorReset)
}

// GetTimerStateColor retorna el color apropiado para el estado del timer
func GetTimerStateColor(state string) Color {
	switch state {
	case "TRABAJO":
		return ColorRed
	case "DESCANSO":
		return ColorCyan
	case "DESCANSO LARGO":
		return ColorBlue
	case "PAUSED", "PAUSADO":
		return ColorYellow
	default:
		return ColorWhite
	}
}

// GetStateColor retorna el color apropiado para cada estado (compatibilidad)
func GetStateColor(state string) string {
	color := GetTimerStateColor(state)
	return string(color) + string(ColorBold)
}

// GetEfficiencyColor retorna color basado en el porcentaje de eficiencia
func GetEfficiencyColor(efficiency float64) Color {
	switch {
	case efficiency >= 90:
		return ColorBrightGreen
	case efficiency >= 75:
		return ColorGreen
	case efficiency >= 60:
		return ColorYellow
	case efficiency >= 40:
		return ColorOrange
	default:
		return ColorRed
	}
}

// GetStreakColor retorna color para la racha actual
func GetStreakColor(streak int) Color {
	switch {
	case streak >= 10:
		return ColorOrange // Fuego intenso
	case streak >= 5:
		return ColorBrightRed // Fuego medio
	case streak >= 3:
		return ColorRed // Fuego inicial
	case streak >= 1:
		return ColorYellow // Chispa
	default:
		return ColorGray // Sin racha
	}
}

// Theme representa un tema de colores
type Theme struct {
	Name       string
	Primary    Color
	Secondary  Color
	Success    Color
	Warning    Color
	Error      Color
	Info       Color
	Text       Color
	Background Color
}

// Temas predefinidos
var (
	// Tema clásico Pomodoro (rojo/tomate)
	ClassicTheme = Theme{
		Name:       "Clásico",
		Primary:    ColorRed,
		Secondary:  ColorOrange,
		Success:    ColorGreen,
		Warning:    ColorYellow,
		Error:      ColorBrightRed,
		Info:       ColorCyan,
		Text:       ColorWhite,
		Background: ColorBlack,
	}

	// Tema océano (azules y verdes)
	OceanTheme = Theme{
		Name:       "Océano",
		Primary:    ColorBlue,
		Secondary:  ColorCyan,
		Success:    ColorGreen,
		Warning:    ColorYellow,
		Error:      ColorRed,
		Info:       ColorBrightCyan,
		Text:       ColorWhite,
		Background: ColorBlack,
	}

	// Tema bosque (verdes)
	ForestTheme = Theme{
		Name:       "Bosque",
		Primary:    ColorGreen,
		Secondary:  ColorBrightGreen,
		Success:    ColorBrightGreen,
		Warning:    ColorYellow,
		Error:      ColorRed,
		Info:       ColorCyan,
		Text:       ColorWhite,
		Background: ColorBlack,
	}

	// Tema monocromo (escala de grises)
	MonoTheme = Theme{
		Name:       "Monocromo",
		Primary:    ColorWhite,
		Secondary:  ColorGray,
		Success:    ColorBrightWhite,
		Warning:    ColorGray,
		Error:      ColorWhite,
		Info:       ColorGray,
		Text:       ColorWhite,
		Background: ColorBlack,
	}
)

// GetAvailableThemes retorna lista de temas disponibles
func GetAvailableThemes() []Theme {
	return []Theme{
		ClassicTheme,
		OceanTheme,
		ForestTheme,
		MonoTheme,
	}
}

// ApplyTheme aplica colores del tema a un texto según el tipo
func (t Theme) ApplyTheme(text string, colorType string, useColors bool) string {
	if !useColors {
		return text
	}

	var color Color
	switch colorType {
	case "primary":
		color = t.Primary
	case "secondary":
		color = t.Secondary
	case "success":
		color = t.Success
	case "warning":
		color = t.Warning
	case "error":
		color = t.Error
	case "info":
		color = t.Info
	case "text":
		color = t.Text
	default:
		color = t.Text
	}

	return Colorize(text, color, useColors)
}

// ProgressBarStyle define estilos para barras de progreso
type ProgressBarStyle struct {
	FilledChar  string
	EmptyChar   string
	FilledColor Color
	EmptyColor  Color
	BorderColor Color
}

// Estilos de barra de progreso predefinidos
var (
	ClassicProgressBar = ProgressBarStyle{
		FilledChar:  "█",
		EmptyChar:   "░",
		FilledColor: ColorGreen,
		EmptyColor:  ColorGray,
		BorderColor: ColorWhite,
	}

	MinimalProgressBar = ProgressBarStyle{
		FilledChar:  "▰",
		EmptyChar:   "▱",
		FilledColor: ColorCyan,
		EmptyColor:  ColorGray,
		BorderColor: ColorWhite,
	}

	RetroProgressBar = ProgressBarStyle{
		FilledChar:  "▓",
		EmptyChar:   "▒",
		FilledColor: ColorGreen,
		EmptyColor:  ColorGray,
		BorderColor: ColorYellow,
	}
)

// CreateStyledProgressBar crea una barra de progreso con estilo personalizado
func CreateStyledProgressBar(progress float64, width int, style ProgressBarStyle, useColors bool) string {
	filled := int(progress * float64(width))
	empty := width - filled

	filledPart := Colorize(strings.Repeat(style.FilledChar, filled), style.FilledColor, useColors)
	emptyPart := Colorize(strings.Repeat(style.EmptyChar, empty), style.EmptyColor, useColors)

	brackets := Colorize("[", style.BorderColor, useColors) + filledPart + emptyPart + Colorize("]", style.BorderColor, useColors)

	return brackets
}

// IsColorSupported verifica si la terminal soporta colores
func IsColorSupported() bool {
	// Verificación básica de soporte de colores
	// En una implementación más avanzada, podrías verificar variables de entorno
	// como TERM, COLORTERM, etc.
	return true // Por simplicidad, asumimos soporte
}
