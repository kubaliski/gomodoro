package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
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

// GetProgressBar crea una barra de progreso simple
func GetProgressBar(remaining, total time.Duration, width int) string {
	if total <= 0 {
		return strings.Repeat("█", width)
	}

	progress := float64(total-remaining) / float64(total)
	filled := int(progress * float64(width))

	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return bar
}

// DisplayTimer muestra el estado actual del timer
func DisplayTimer(remaining time.Duration, state string) {
	ClearScreen()

	// Determinar duración total basada en el estado
	var totalDuration time.Duration
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

	// ASCII Art Header
	fmt.Println("+================================+")
	fmt.Println("|          POMODORO CLI          |")
	fmt.Println("+================================+")
	fmt.Println("=====================================")

	// Estado y tiempo
	fmt.Printf(">> Estado: %s\n", state)
	fmt.Printf(">> Tiempo restante: %s\n", FormatDuration(remaining))

	// Barra de progreso
	progressBar := GetProgressBar(remaining, totalDuration, 30)
	fmt.Printf(">> Progreso: [%s]\n", progressBar)

	// Porcentaje
	progress := float64(totalDuration-remaining) / float64(totalDuration) * 100
	if progress > 100 {
		progress = 100
	}
	if progress < 0 {
		progress = 0
	}
	fmt.Printf(">> Completado: %.1f%%\n", progress)

	fmt.Println("=====================================")
	fmt.Println(">> Presiona Ctrl+C para salir")
}
