package main

import (
	"fmt"
	"os"

	"github.com/kubaliski/pomodoro-cli/internal/cli"
	"github.com/kubaliski/pomodoro-cli/internal/session"
)

func main() {
	// Parsear argumentos de línea de comandos
	args := cli.ParseArgs()

	// Mostrar ayuda si se solicita
	if args.Help {
		cli.ShowHelp()
		return
	}

	// Validar argumentos y crear configuración
	config, err := cli.ValidateAndCreateConfig(args)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	// Crear y ejecutar sesión
	pomodoroSession := session.NewSession(config)
	pomodoroSession.Run()
}
