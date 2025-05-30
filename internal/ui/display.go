package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Códigos de color ANSI
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"

	// Colores de fondo
	BgRed    = "\033[41m"
	BgGreen  = "\033[42m"
	BgYellow = "\033[43m"
	BgBlue   = "\033[44m"
	BgPurple = "\033[45m"
	BgCyan   = "\033[46m"
)

var isFirstDisplay = true

// ClearScreen limpia la pantalla de la terminal
func ClearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

// FormatDuration convierte una duración a formato MM:SS
func FormatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	if totalSeconds < 0 {
		totalSeconds = 0
	}

	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// GetProgressBar crea una barra de progreso simple con colores
func GetProgressBar(remaining, total time.Duration, width int) string {
	if total <= 0 {
		return ColorGreen + strings.Repeat("█", width) + ColorReset
	}

	progress := float64(total-remaining) / float64(total)
	filled := int(progress * float64(width))

	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	// Cambiar color según el progreso
	var barColor string
	if progress < 0.5 {
		barColor = ColorRed
	} else if progress < 0.8 {
		barColor = ColorYellow
	} else {
		barColor = ColorGreen
	}

	bar := barColor + strings.Repeat("█", filled) + ColorReset +
		ColorWhite + strings.Repeat("░", width-filled) + ColorReset
	return bar
}

// GetStateColor retorna el color apropiado para cada estado
func GetStateColor(state string) string {
	switch state {
	case "TRABAJO":
		return ColorRed + ColorBold
	case "DESCANSO":
		return ColorGreen + ColorBold
	case "DESCANSO LARGO":
		return ColorBlue + ColorBold
	default:
		return ColorWhite + ColorBold
	}
}

// DisplayTimer muestra el estado actual del timer (versión simple sin movimiento de cursor)
func DisplayTimer(remaining time.Duration, state string, totalDuration ...time.Duration) {
	// Determinar duración total
	var total time.Duration
	if len(totalDuration) > 0 {
		total = totalDuration[0]
	} else {
		// Fallback a valores por defecto si no se proporciona
		switch state {
		case "TRABAJO":
			total = 25 * time.Minute
		case "DESCANSO":
			total = 5 * time.Minute
		case "DESCANSO LARGO":
			total = 15 * time.Minute
		default:
			total = 25 * time.Minute
		}
	}

	// Solo mostrar header la primera vez
	if isFirstDisplay {
		ClearScreen()
		// ASCII Art Header simple con color
		fmt.Print(ColorCyan + ColorBold)
		fmt.Println("+================================+")
		fmt.Println("|          POMODORO CLI          |")
		fmt.Println("+================================+")
		fmt.Print(ColorReset)
		fmt.Println()
		isFirstDisplay = false
	}

	// Información dinámica en una sola línea que se sobrescribe
	stateColor := GetStateColor(state)

	// Tiempo restante con color dinámico
	timeColor := ColorWhite
	if remaining < 5*time.Minute {
		timeColor = ColorRed + ColorBold
	} else if remaining < 10*time.Minute {
		timeColor = ColorYellow
	}

	// Barra de progreso
	progressBar := GetProgressBar(remaining, total, 20) // Más corta para una línea

	// Porcentaje
	progress := float64(total-remaining) / float64(total) * 100
	if progress > 100 {
		progress = 100
	}
	if progress < 0 {
		progress = 0
	}

	var percentColor string
	if progress < 25 {
		percentColor = ColorRed
	} else if progress < 75 {
		percentColor = ColorYellow
	} else {
		percentColor = ColorGreen
	}

	// Una sola línea que se actualiza con \r (carriage return)
	fmt.Printf("\r%s%s%s | %s%s%s | [%s] %s%.1f%%%s | Ctrl+C para salir",
		stateColor, state, ColorReset,
		timeColor, FormatDuration(remaining), ColorReset,
		progressBar,
		percentColor, progress, ColorReset)
}

// ResetDisplay reinicia el estado del display (para usar entre sesiones)
func ResetDisplay() {
	isFirstDisplay = true
	fmt.Println() // Nueva línea después del timer en línea
}
