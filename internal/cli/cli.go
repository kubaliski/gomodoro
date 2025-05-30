package cli

import (
	"flag"
	"fmt"
	"time"

	"github.com/kubaliski/pomodoro-cli/internal/config"
)

// Args representa los argumentos de línea de comandos
type Args struct {
	WorkDuration string
	ShortBreak   string
	LongBreak    string
	Help         bool
}

// ParseArgs parsea los argumentos de línea de comandos
func ParseArgs() *Args {
	args := &Args{}

	flag.StringVar(&args.WorkDuration, "work", "25m", "Duración de la sesión de trabajo (ej: 25m, 30m)")
	flag.StringVar(&args.ShortBreak, "break", "5m", "Duración del descanso corto (ej: 5m, 10m)")
	flag.StringVar(&args.LongBreak, "long-break", "15m", "Duración del descanso largo (ej: 15m, 20m)")
	flag.BoolVar(&args.Help, "help", false, "Mostrar ayuda")

	flag.Parse()

	return args
}

// ShowHelp muestra la ayuda del programa
func ShowHelp() {
	fmt.Println("+================================+")
	fmt.Println("|          POMODORO CLI          |")
	fmt.Println("+================================+")
	fmt.Println()
	fmt.Println("Un temporizador Pomodoro completo con ciclos automáticos.")
	fmt.Println()
	fmt.Println("Uso:")
	fmt.Println("  pomodoro [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -work string")
	fmt.Println("        Duración de la sesión de trabajo (default \"25m\")")
	fmt.Println("  -break string")
	fmt.Println("        Duración del descanso corto (default \"5m\")")
	fmt.Println("  -long-break string")
	fmt.Println("        Duración del descanso largo (default \"15m\")")
	fmt.Println("  -help")
	fmt.Println("        Mostrar esta ayuda")
	fmt.Println()
	fmt.Println("Ejemplos:")
	fmt.Println("  pomodoro                    # Configuración estándar (25m/5m/15m)")
	fmt.Println("  pomodoro -work=30m          # Sesiones de 30 minutos")
	fmt.Println("  pomodoro -work=45m -break=10m -long-break=20m")
	fmt.Println("  pomodoro -work=5s -break=3s # Para pruebas rápidas")
	fmt.Println()
	fmt.Println("Funcionamiento:")
	fmt.Println("  • Alterna automáticamente entre trabajo y descansos")
	fmt.Println("  • Descanso largo cada 4 pomodoros completados")
	fmt.Println("  • Usa Ctrl+C para salir en cualquier momento")
}

// ValidateAndCreateConfig valida los argumentos y crea la configuración
func ValidateAndCreateConfig(args *Args) (*config.Config, error) {
	// Parsear duración de trabajo
	workDur, err := time.ParseDuration(args.WorkDuration)
	if err != nil {
		return nil, fmt.Errorf("duración de trabajo inválida '%s'. Usa formato como 25m, 30m, etc", args.WorkDuration)
	}

	// Parsear duración de descanso corto
	shortDur, err := time.ParseDuration(args.ShortBreak)
	if err != nil {
		return nil, fmt.Errorf("duración de descanso inválida '%s'. Usa formato como 5m, 10m, etc", args.ShortBreak)
	}

	// Parsear duración de descanso largo
	longDur, err := time.ParseDuration(args.LongBreak)
	if err != nil {
		return nil, fmt.Errorf("duración de descanso largo inválida '%s'. Usa formato como 15m, 20m, etc", args.LongBreak)
	}

	// Validaciones adicionales
	if workDur <= 0 {
		return nil, fmt.Errorf("la duración de trabajo debe ser mayor que 0")
	}

	if shortDur <= 0 {
		return nil, fmt.Errorf("la duración de descanso debe ser mayor que 0")
	}

	if longDur <= 0 {
		return nil, fmt.Errorf("la duración de descanso largo debe ser mayor que 0")
	}

	// Crear configuración
	cfg := &config.Config{
		WorkDuration:      workDur,
		ShortBreak:        shortDur,
		LongBreak:         longDur,
		LongBreakInterval: 4,
	}

	return cfg, nil
}
