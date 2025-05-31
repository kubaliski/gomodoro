package handlers

import (
	"bufio"
	"os"
	"strings"
)

// InputManager maneja la entrada de comandos del usuario
type InputManager struct {
	reader    *bufio.Reader
	inputChan chan string
}

// NewInputManager crea un nuevo gestor de input
func NewInputManager() *InputManager {
	return &InputManager{
		reader:    bufio.NewReader(os.Stdin),
		inputChan: make(chan string, 10),
	}
}

// StartListener inicia el listener de input en background
func (im *InputManager) StartListener() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := strings.TrimSpace(strings.ToLower(scanner.Text()))
		select {
		case im.inputChan <- input:
		default:
			// Canal lleno, ignorar entrada
		}
	}
}

// HandleInput procesa los comandos usando el procesador proporcionado
func (im *InputManager) HandleInput(processor func(string)) {
	for input := range im.inputChan {
		processor(input)
	}
}

// GetInputChannel retorna el canal de input (para casos especiales)
func (im *InputManager) GetInputChannel() chan string {
	return im.inputChan
}
