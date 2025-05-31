package input

import (
	"bufio"
	"os"
	"strings"
)

// Command representa un comando del usuario
type Command string

const (
	CommandPause   Command = "p"
	CommandResume  Command = "r"
	CommandSkip    Command = "s"
	CommandQuit    Command = "q"
	CommandHelp    Command = "h"
	CommandUnknown Command = "unknown"
)

// StartListening inicia la escucha de comandos del usuario en una goroutine
func StartListening(commandChan chan<- Command) {
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := strings.TrimSpace(strings.ToLower(scanner.Text()))

			var cmd Command
			switch input {
			case "p", "pause":
				cmd = CommandPause
			case "r", "resume":
				cmd = CommandResume
			case "s", "skip":
				cmd = CommandSkip
			case "q", "quit":
				cmd = CommandQuit
			case "h", "help":
				cmd = CommandHelp
			case "":
				// Ignorar líneas vacías
				continue
			default:
				cmd = CommandUnknown
			}

			// Enviar comando al canal
			select {
			case commandChan <- cmd:
			default:
				// Canal lleno, ignorar comando
			}
		}
	}()
}

// ShowControls muestra los controles disponibles
func ShowControls() {
	return // Los mostraremos en el display principal
}
