package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Variables de estado del display
var (
	headerShown        = false
	lastDisplayContent = ""
)

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

// GetProgressBar crea una barra de progreso simple con colores usando el nuevo sistema
func GetProgressBar(remaining, total time.Duration, width int) string {
	if total <= 0 {
		return Colorize(strings.Repeat("█", width), ColorGreen, true)
	}

	progress := float64(total-remaining) / float64(total)
	filled := int(progress * float64(width))

	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	// Cambiar color según el progreso usando el nuevo sistema
	var barColor Color
	if progress < 0.5 {
		barColor = ColorRed
	} else if progress < 0.8 {
		barColor = ColorYellow
	} else {
		barColor = ColorGreen
	}

	bar := Colorize(strings.Repeat("█", filled), barColor, true) +
		Colorize(strings.Repeat("░", width-filled), ColorWhite, true)
	return bar
}

// DisplayTimer muestra el estado actual del timer (versión anti-parpadeo mejorada)
func DisplayTimer(remaining time.Duration, state string, args ...interface{}) {
	var timerStatus string
	var totalDuration time.Duration

	// Procesar argumentos variables para máxima compatibilidad
	if len(args) == 2 {
		if status, ok := args[0].(string); ok {
			timerStatus = status
		} else {
			timerStatus = "RUNNING"
		}

		if duration, ok := args[1].(time.Duration); ok {
			totalDuration = duration
		} else {
			totalDuration = 25 * time.Minute
		}
	} else if len(args) == 1 {
		timerStatus = "RUNNING"
		if duration, ok := args[0].(time.Duration); ok {
			totalDuration = duration
		} else {
			totalDuration = 25 * time.Minute
		}
	} else {
		timerStatus = "RUNNING"
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
		fmt.Print(Colorize("+================================+", ColorCyan, true))
		fmt.Println()
		fmt.Print(Colorize("|          GOMODORO CLI          |", ColorCyan, true))
		fmt.Println()
		fmt.Print(Colorize("+================================+", ColorCyan, true))
		fmt.Println()
		fmt.Println()
		fmt.Println("Escribe comandos: (p)ausar (r)eanudar (s)altar (q)salir (h)ayuda")
		fmt.Println()
		headerShown = true
	}

	// Calcular información para mostrar usando el nuevo sistema de colores
	stateColor := GetTimerStateColor(state)

	var statusColor Color
	switch strings.ToUpper(timerStatus) {
	case "PAUSED", "PAUSADO":
		statusColor = ColorYellow
	case "RUNNING", "CORRIENDO":
		statusColor = ColorGreen
	case "STOPPED", "DETENIDO":
		statusColor = ColorRed
	default:
		statusColor = ColorWhite
	}

	timeColor := ColorWhite
	if remaining < 5*time.Minute && strings.ToUpper(timerStatus) == "RUNNING" {
		timeColor = ColorRed
	} else if remaining < 10*time.Minute && strings.ToUpper(timerStatus) == "RUNNING" {
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

	var percentColor Color
	if progress < 25 {
		percentColor = ColorRed
	} else if progress < 75 {
		percentColor = ColorYellow
	} else {
		percentColor = ColorGreen
	}

	// Construir el contenido completo usando el nuevo sistema
	content := fmt.Sprintf("%s | %s | %s | [%s] %s",
		Colorize(state, stateColor, true),
		Colorize(timerStatus, statusColor, true),
		Colorize(FormatDuration(remaining), timeColor, true),
		progressBar,
		Colorize(fmt.Sprintf("%.1f%%", progress), percentColor, true))

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

// ShowHeader muestra el header principal del CLI
func ShowHeader(title string) {
	ClearScreen()
	fmt.Println(Colorize("╔══════════════════════════════════════════════════════════════╗", ColorCyan, true))
	fmt.Printf(Colorize("║%s║", ColorCyan, true)+"\n", centerText(title, 62))
	fmt.Println(Colorize("╚══════════════════════════════════════════════════════════════╝", ColorCyan, true))
	fmt.Println()
}

// centerText centra texto dentro de un ancho específico
func centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width-3] + "..."
	}

	padding := width - len(text)
	leftPad := padding / 2
	rightPad := padding - leftPad

	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

// ShowSection muestra una sección con título
func ShowSection(title string, color Color) {
	fmt.Println(Colorize(title, color, true))
	fmt.Println(Colorize(strings.Repeat("─", len(title)), ColorGray, true))
}

// ShowSeparator muestra una línea separadora
func ShowSeparator(width int) {
	fmt.Println(Colorize(strings.Repeat("─", width), ColorGray, true))
}

// ShowBox muestra texto en una caja
func ShowBox(title, content string, color Color) {
	lines := strings.Split(content, "\n")
	maxWidth := len(title)

	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	boxWidth := maxWidth + 4

	// Top border
	fmt.Println(Colorize("┌"+strings.Repeat("─", boxWidth-2)+"┐", color, true))

	// Title
	fmt.Printf(Colorize("│ %s%s │", color, true)+"\n",
		title, strings.Repeat(" ", boxWidth-4-len(title)))

	// Separator
	fmt.Println(Colorize("├"+strings.Repeat("─", boxWidth-2)+"┤", color, true))

	// Content
	for _, line := range lines {
		fmt.Printf(Colorize("│ %s%s │", color, true)+"\n",
			line, strings.Repeat(" ", boxWidth-4-len(line)))
	}

	// Bottom border
	fmt.Println(Colorize("└"+strings.Repeat("─", boxWidth-2)+"┘", color, true))
}
