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

var headerShown = false
var lastDisplayContent = ""

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

// DisplayTimer muestra el estado actual del timer (versión anti-parpadeo)
func DisplayTimer(remaining time.Duration, state string, args ...interface{}) {
	var timerStatus string
	var totalDuration time.Duration

	// Procesar argumentos variables para máxima compatibilidad
	if len(args) == 2 {
		if status, ok := args[0].(string); ok {
			timerStatus = status
		} else {
			timerStatus = "CORRIENDO"
		}

		if duration, ok := args[1].(time.Duration); ok {
			totalDuration = duration
		} else {
			totalDuration = 25 * time.Minute
		}
	} else if len(args) == 1 {
		timerStatus = "CORRIENDO"
		if duration, ok := args[0].(time.Duration); ok {
			totalDuration = duration
		} else {
			totalDuration = 25 * time.Minute
		}
	} else {
		timerStatus = "CORRIENDO"
		switch state {
		case "TRABAJO":
			totalDuration = 25 * time.Minute
		case "DESCANSO":
			totalDuration = 5 * time.Minute
		case "DESCANSO LARGO":
			totalDuration = 15 * time.Minute
		default:
			totalDuration = 25 * time.Minute
		}
	}

	// Mostrar header una sola vez
	if !headerShown {
		ClearScreen()
		fmt.Print(ColorCyan + ColorBold)
		fmt.Println("+================================+")
		fmt.Println("|          POMODORO CLI          |")
		fmt.Println("+================================+")
		fmt.Print(ColorReset)
		fmt.Println()
		fmt.Println("Escribe comandos: (p)ausar (r)eanudar (s)altar (q)salir (h)ayuda")
		fmt.Println()
		headerShown = true
	}

	// Calcular información para mostrar
	stateColor := GetStateColor(state)

	var statusColor string
	switch timerStatus {
	case "PAUSADO":
		statusColor = ColorYellow + ColorBold
	case "CORRIENDO":
		statusColor = ColorGreen
	case "DETENIDO":
		statusColor = ColorRed
	default:
		statusColor = ColorWhite
	}

	timeColor := ColorWhite
	if remaining < 5*time.Minute && timerStatus == "CORRIENDO" {
		timeColor = ColorRed + ColorBold
	} else if remaining < 10*time.Minute && timerStatus == "CORRIENDO" {
		timeColor = ColorYellow
	}

	progressBar := GetProgressBar(remaining, totalDuration, 20)

	progress := float64(totalDuration-remaining) / float64(totalDuration) * 100
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

	// Construir el contenido completo
	content := fmt.Sprintf("%s%s%s | %s%s%s | %s%s%s | [%s] %s%.1f%%%s",
		stateColor, state, ColorReset,
		statusColor, timerStatus, ColorReset,
		timeColor, FormatDuration(remaining), ColorReset,
		progressBar,
		percentColor, progress, ColorReset)

	// Solo actualizar si hay cambios significativos (evitar parpadeo)
	if content != lastDisplayContent {
		fmt.Print(content)
		lastDisplayContent = content
	} else {
		// Si no hay cambios, solo imprimir el contenido sin limpiar
		fmt.Print(content)
	}
}

// DisplayTimerWithPrompt muestra el timer con un prompt para comandos
func DisplayTimerWithPrompt(remaining time.Duration, state string, args ...interface{}) {
	DisplayTimer(remaining, state, args...)
	fmt.Print(" > ")
}

// ResetDisplay reinicia el estado del display (para usar entre sesiones)
func ResetDisplay() {
	headerShown = false
	lastDisplayContent = ""
	fmt.Println() // Nueva línea para separar sesiones
}
