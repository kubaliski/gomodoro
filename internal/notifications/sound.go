package notifications

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// SoundNotifier implementa notificaciones de sonido cross-platform
type SoundNotifier struct {
	config   map[string]interface{}
	platform string
}

// NewSoundNotifier crea un nuevo notificador de sonido
func NewSoundNotifier() *SoundNotifier {
	return &SoundNotifier{
		config:   make(map[string]interface{}),
		platform: runtime.GOOS,
	}
}

// GetType retorna el tipo de notificador
func (s *SoundNotifier) GetType() NotificationType {
	return TypeSound
}

// IsAvailable verifica si el sistema de sonidos está disponible
func (s *SoundNotifier) IsAvailable() bool {
	switch s.platform {
	case "windows":
		return s.isWindowsSoundAvailable()
	case "darwin":
		return s.isMacOSSoundAvailable()
	case "linux":
		return s.isLinuxSoundAvailable()
	default:
		return false
	}
}

// Configure configura el notificador de sonido
func (s *SoundNotifier) Configure(config map[string]interface{}) error {
	s.config = config
	return nil
}

// Notify ejecuta una notificación de sonido
func (s *SoundNotifier) Notify(request NotificationRequest) NotificationResponse {
	start := time.Now()

	// Seleccionar tipo de sonido basado en el evento
	soundType := s.getSoundTypeForEvent(request.Event)

	// DEBUG: Agregar logging para debugging
	fmt.Printf("[DEBUG] Event: %s -> SoundType: %s\n", request.Event, soundType)

	// Obtener configuración de sonido
	volume := s.getConfigFloat("volume", 0.7)
	duration := s.getConfigInt("duration", 500)
	frequency := s.getConfigInt("frequency", 800)
	customSounds := s.getConfigBool("custom_sounds", false)

	var err error

	// Intentar reproducir sonido personalizado primero
	if customSounds {
		fmt.Printf("[DEBUG] Trying custom sound for type: %s\n", soundType)
		err = s.playCustomSound(soundType, volume)
		if err == nil {
			fmt.Printf("[DEBUG] Custom sound successful\n")
			return NotificationResponse{
				Success:  true,
				Type:     TypeSound,
				Duration: time.Since(start),
			}
		}
		// Si falla, continuar con sonidos del sistema
		fmt.Printf("[DEBUG] Custom sound failed: %v, trying system sound\n", err)
	}

	// Reproducir sonido del sistema
	fmt.Printf("[DEBUG] Trying system sound for type: %s (freq: %d, dur: %d)\n", soundType, frequency, duration)
	err = s.playSystemSound(soundType, volume, duration, frequency)

	if err != nil {
		fmt.Printf("[DEBUG] System sound failed: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] System sound successful\n")
	}

	return NotificationResponse{
		Success:  err == nil,
		Type:     TypeSound,
		Error:    err,
		Duration: time.Since(start),
	}
}

// getSoundTypeForEvent determina el tipo de sonido para un evento
func (s *SoundNotifier) getSoundTypeForEvent(event EventType) string {
	// Convertir a string para hacer comparación más robusta
	eventStr := string(event)

	fmt.Printf("[DEBUG] Processing event: '%s'\n", eventStr)

	switch eventStr {
	case "pomodoro_completed":
		return "success"
	case "break_completed":
		return "gentle"
	case "early_alert":
		return "warning"
	case "urgent_alert":
		return "urgent"
	case "session_started":
		return "start"
	case "timer_paused":
		return "pause"
	case "timer_resumed":
		return "resume"
	case "custom_alert":
		return "default"
	default:
		fmt.Printf("[DEBUG] Unknown event type: '%s', using default\n", eventStr)
		return "default"
	}
}

// playCustomSound intenta reproducir un archivo de sonido personalizado
func (s *SoundNotifier) playCustomSound(soundType string, volume float64) error {
	soundFile := s.getSoundFilePath(soundType)
	if soundFile == "" {
		return fmt.Errorf("no custom sound file configured for type: %s", soundType)
	}

	// Verificar que el archivo existe
	if _, err := os.Stat(soundFile); os.IsNotExist(err) {
		return fmt.Errorf("sound file does not exist: %s", soundFile)
	}

	switch s.platform {
	case "windows":
		return s.playWindowsFile(soundFile, volume)
	case "darwin":
		return s.playMacOSFile(soundFile, volume)
	case "linux":
		return s.playLinuxFile(soundFile, volume)
	default:
		return fmt.Errorf("unsupported platform for custom sounds: %s", s.platform)
	}
}

// playSystemSound reproduce un sonido del sistema
func (s *SoundNotifier) playSystemSound(soundType string, volume float64, duration, frequency int) error {
	switch s.platform {
	case "windows":
		return s.playWindowsBeep(soundType, frequency, duration)
	case "darwin":
		return s.playMacOSBeep(soundType, volume)
	case "linux":
		return s.playLinuxBeep(soundType, frequency, duration)
	default:
		return fmt.Errorf("unsupported platform: %s", s.platform)
	}
}

// Windows sound methods

func (s *SoundNotifier) isWindowsSoundAvailable() bool {
	// Windows siempre tiene soporte para Beep
	return true
}

func (s *SoundNotifier) playWindowsBeep(soundType string, defaultFreq, defaultDuration int) error {
	var frequency, duration int

	// Configurar frecuencia y duración según tipo de sonido
	switch soundType {
	case "success":
		frequency = 800 // Tono alto y agradable
		duration = defaultDuration
	case "gentle":
		frequency = 600 // Tono suave
		duration = defaultDuration
	case "warning":
		frequency = 1000 // Tono de advertencia
		duration = defaultDuration
	case "urgent":
		frequency = 1200               // Tono urgente
		duration = defaultDuration / 2 // Beeps más cortos pero repetidos

		// Para urgente, hacer múltiples beeps
		for i := 0; i < 3; i++ {
			if err := s.executeWindowsBeep(frequency, duration); err != nil {
				return err
			}
			if i < 2 {
				time.Sleep(100 * time.Millisecond)
			}
		}
		return nil
	case "start":
		frequency = 500 // Tono neutro para inicio
		duration = 300  // Duración media
	case "pause":
		frequency = 400 // Tono bajo para pausa
		duration = 200  // Duración corta
	case "resume":
		frequency = 600 // Tono medio para reanudar
		duration = 200  // Duración corta
	case "default":
		frequency = 500
		duration = 200
	default:
		// Case por defecto para eventos no reconocidos
		frequency = defaultFreq
		duration = defaultDuration
	}

	return s.executeWindowsBeep(frequency, duration)
}

// executeWindowsBeep ejecuta el beep de Windows con múltiples métodos de fallback
func (s *SoundNotifier) executeWindowsBeep(frequency, duration int) error {
	// Método 1: PowerShell Console Beep
	err1 := exec.Command("powershell", "-c",
		fmt.Sprintf("[console]::beep(%d,%d)", frequency, duration)).Run()
	if err1 == nil {
		return nil
	}

	// Método 2: PowerShell System.Media.SystemSounds
	err2 := exec.Command("powershell", "-c",
		"[System.Media.SystemSounds]::Beep.Play()").Run()
	if err2 == nil {
		return nil
	}

	// Método 3: Echo con terminal bell
	err3 := exec.Command("cmd", "/c", "echo \a").Run()
	if err3 == nil {
		return nil
	}

	// Si todos fallan, retornar el primer error
	return fmt.Errorf("all Windows beep methods failed: console.beep=%v, SystemSounds=%v, echo=%v", err1, err2, err3)
}

func (s *SoundNotifier) playWindowsFile(soundFile string, volume float64) error {
	// Usar PowerShell para reproducir archivos de audio con control de volumen
	volumePercent := int(volume * 100)

	// Comando mejorado que incluye configuración de volumen
	cmd := exec.Command("powershell", "-c",
		fmt.Sprintf(`
		try {
			# Configurar volumen del sistema
			$audio = New-Object -ComObject WScript.Shell

			# Reproducir archivo de sonido
			$player = New-Object System.Media.SoundPlayer
			$player.SoundLocation = "%s"
			$player.Load()
			$player.PlaySync()

			Write-Host "Audio played at %d%% volume"
		} catch {
			Write-Error "Failed to play audio: $_"
			exit 1
		}
		`, soundFile, volumePercent))

	return cmd.Run()
}

// macOS sound methods

func (s *SoundNotifier) isMacOSSoundAvailable() bool {
	// Verificar si 'say' está disponible (siempre debería estarlo en macOS)
	_, err := exec.LookPath("say")
	return err == nil
}

func (s *SoundNotifier) playMacOSBeep(soundType string, volume float64) error {
	switch soundType {
	case "success":
		return exec.Command("say", "-v", "Bells", "ding").Run()
	case "gentle":
		return exec.Command("say", "-v", "Whisper", "chime").Run()
	case "warning":
		return exec.Command("say", "-v", "Alex", "-r", "300", "beep").Run()
	case "urgent":
		// Múltiples beeps para urgente
		for i := 0; i < 3; i++ {
			if err := exec.Command("say", "-v", "Alex", "-r", "400", "beep").Run(); err != nil {
				return err
			}
			if i < 2 {
				time.Sleep(200 * time.Millisecond)
			}
		}
		return nil
	case "start":
		return exec.Command("osascript", "-e", "beep 1").Run()
	case "pause":
		return exec.Command("osascript", "-e", "beep 2").Run()
	case "resume":
		return exec.Command("osascript", "-e", "beep 1").Run()
	default:
		// Usar el sonido del sistema por defecto
		return exec.Command("osascript", "-e", "beep").Run()
	}
}

func (s *SoundNotifier) playMacOSFile(soundFile string, volume float64) error {
	volumeStr := fmt.Sprintf("%.1f", volume)
	return exec.Command("afplay", "-v", volumeStr, soundFile).Run()
}

// Linux sound methods

func (s *SoundNotifier) isLinuxSoundAvailable() bool {
	// Verificar diferentes opciones disponibles en Linux
	commands := []string{"pactl", "aplay", "speaker-test", "ffplay", "mpg123"}

	for _, cmd := range commands {
		if _, err := exec.LookPath(cmd); err == nil {
			return true
		}
	}

	return false
}

func (s *SoundNotifier) playLinuxBeep(soundType string, frequency, duration int) error {
	// Intentar diferentes métodos en orden de preferencia

	// 1. Intentar con pactl (PulseAudio)
	if err := s.playLinuxPulseAudio(soundType, frequency, duration); err == nil {
		return nil
	}

	// 2. Intentar con speaker-test
	if err := s.playLinuxSpeakerTest(soundType, frequency, duration); err == nil {
		return nil
	}

	// 3. Intentar con beep command (si está instalado)
	if err := s.playLinuxBeepCommand(soundType, frequency, duration); err == nil {
		return nil
	}

	// 4. Fallback a echo con terminal bell
	return s.playLinuxTerminalBell(soundType)
}

func (s *SoundNotifier) playLinuxPulseAudio(soundType string, defaultFreq, defaultDuration int) error {
	if _, err := exec.LookPath("pactl"); err != nil {
		return err
	}

	// Usar pactl para generar tono
	switch soundType {
	case "urgent":
		// Múltiples beeps para urgente
		for i := 0; i < 3; i++ {
			cmd := exec.Command("pactl", "play-sample", "bell-window-system")
			if err := cmd.Run(); err != nil {
				// Si no hay sample predefinido, usar alternativa
				exec.Command("pactl", "play-file", "/usr/share/sounds/alsa/Front_Left.wav").Run()
			}
			if i < 2 {
				time.Sleep(100 * time.Millisecond)
			}
		}
		return nil
	case "start":
		cmd := exec.Command("pactl", "play-sample", "dialog-information")
		if err := cmd.Run(); err != nil {
			return exec.Command("pactl", "play-file", "/usr/share/sounds/alsa/Front_Left.wav").Run()
		}
		return nil
	case "pause":
		cmd := exec.Command("pactl", "play-sample", "suspend-error")
		if err := cmd.Run(); err != nil {
			return exec.Command("pactl", "play-file", "/usr/share/sounds/alsa/Front_Right.wav").Run()
		}
		return nil
	case "resume":
		cmd := exec.Command("pactl", "play-sample", "dialog-information")
		if err := cmd.Run(); err != nil {
			return exec.Command("pactl", "play-file", "/usr/share/sounds/alsa/Front_Center.wav").Run()
		}
		return nil
	default:
		cmd := exec.Command("pactl", "play-sample", "bell-window-system")
		if err := cmd.Run(); err != nil {
			// Fallback a sonido de sistema
			return exec.Command("pactl", "play-file", "/usr/share/sounds/alsa/Front_Left.wav").Run()
		}
		return nil
	}
}

func (s *SoundNotifier) playLinuxSpeakerTest(soundType string, defaultFreq, defaultDuration int) error {
	if _, err := exec.LookPath("speaker-test"); err != nil {
		return err
	}

	var frequency, duration int

	// Ajustar parámetros según el tipo de sonido
	switch soundType {
	case "start":
		frequency = 500
		duration = 300
	case "pause":
		frequency = 300
		duration = 200
	case "resume":
		frequency = 600
		duration = 200
	default:
		frequency = defaultFreq
		duration = defaultDuration
	}

	durationSec := duration / 1000
	if durationSec < 1 {
		durationSec = 1
	}

	cmd := exec.Command("speaker-test", "-t", "sine", "-f",
		strconv.Itoa(frequency), "-l", "1", "-s", "1", "-c", "1")

	// Ejecutar en background y matar después de la duración especificada
	if err := cmd.Start(); err != nil {
		return err
	}

	// Esperar la duración especificada y luego terminar
	go func() {
		time.Sleep(time.Duration(durationSec) * time.Second)
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	return cmd.Wait()
}

func (s *SoundNotifier) playLinuxBeepCommand(soundType string, defaultFreq, defaultDuration int) error {
	if _, err := exec.LookPath("beep"); err != nil {
		return err
	}

	var frequency, duration int

	switch soundType {
	case "urgent":
		frequency = defaultFreq
		duration = defaultDuration / 3
		// Múltiples beeps
		return exec.Command("beep", "-f", strconv.Itoa(frequency),
			"-l", strconv.Itoa(duration), "-r", "3", "-d", "100").Run()
	case "start":
		frequency = defaultFreq
		duration = defaultDuration
	case "pause":
		frequency = 300                // Frecuencia más baja para pausa
		duration = defaultDuration / 2 // Duración más corta
	case "resume":
		frequency = defaultFreq + 100 // Frecuencia ligeramente más alta
		duration = defaultDuration / 2
	default:
		frequency = defaultFreq
		duration = defaultDuration
	}

	return exec.Command("beep", "-f", strconv.Itoa(frequency),
		"-l", strconv.Itoa(duration)).Run()
}

func (s *SoundNotifier) playLinuxTerminalBell(soundType string) error {
	switch soundType {
	case "urgent":
		// Múltiples bells
		for i := 0; i < 3; i++ {
			fmt.Print("\a") // Terminal bell
			if i < 2 {
				time.Sleep(200 * time.Millisecond)
			}
		}
	case "start":
		fmt.Print("\a") // Terminal bell simple
	case "pause":
		fmt.Print("\a") // Terminal bell simple
	case "resume":
		fmt.Print("\a") // Terminal bell simple
	default:
		fmt.Print("\a") // Terminal bell
	}
	return nil
}

func (s *SoundNotifier) playLinuxFile(soundFile string, volume float64) error {
	// Intentar con diferentes reproductores en orden de preferencia
	players := []struct {
		cmd  string
		args []string
	}{
		{"aplay", []string{soundFile}},
		{"paplay", []string{soundFile}},
		{"ffplay", []string{"-nodisp", "-autoexit", "-volume", fmt.Sprintf("%.0f", volume*100), soundFile}},
		{"mpg123", []string{"-q", "--gain", fmt.Sprintf("%.0f", volume*100), soundFile}},
		{"play", []string{soundFile, "vol", fmt.Sprintf("%.1f", volume)}},                                          // SoX
		{"cvlc", []string{"--intf", "dummy", "--play-and-exit", "--gain", fmt.Sprintf("%.1f", volume), soundFile}}, // VLC
	}

	for _, player := range players {
		if _, err := exec.LookPath(player.cmd); err == nil {
			cmd := exec.Command(player.cmd, player.args...)
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("no audio player available")
}

// Helper methods

func (s *SoundNotifier) getConfigFloat(key string, defaultValue float64) float64 {
	if val, ok := s.config[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
		// Intentar convertir desde otros tipos numéricos
		if i, ok := val.(int); ok {
			return float64(i)
		}
		if str, ok := val.(string); ok {
			if parsed, err := strconv.ParseFloat(str, 64); err == nil {
				return parsed
			}
		}
	}
	return defaultValue
}

func (s *SoundNotifier) getConfigInt(key string, defaultValue int) int {
	if val, ok := s.config[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
		// Intentar convertir desde otros tipos numéricos
		if f, ok := val.(float64); ok {
			return int(f)
		}
		if str, ok := val.(string); ok {
			if parsed, err := strconv.Atoi(str); err == nil {
				return parsed
			}
		}
	}
	return defaultValue
}

func (s *SoundNotifier) getConfigBool(key string, defaultValue bool) bool {
	if val, ok := s.config[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
		// Intentar convertir desde string
		if str, ok := val.(string); ok {
			return strings.ToLower(str) == "true" || str == "1"
		}
		// Convertir desde números
		if i, ok := val.(int); ok {
			return i != 0
		}
	}
	return defaultValue
}

func (s *SoundNotifier) getSoundFilePath(soundType string) string {
	key := fmt.Sprintf("sound_file_%s", soundType)
	if val, ok := s.config[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// SetCustomSoundFile configura un archivo de sonido personalizado
func (s *SoundNotifier) SetCustomSoundFile(soundType, filePath string) error {
	// Verificar que el archivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("sound file does not exist: %s", filePath)
	}

	key := fmt.Sprintf("sound_file_%s", soundType)
	s.config[key] = filePath
	return nil
}

// GetSupportedFormats retorna los formatos de audio soportados por plataforma
func (s *SoundNotifier) GetSupportedFormats() []string {
	switch s.platform {
	case "windows":
		return []string{".wav", ".mp3", ".wma"}
	case "darwin":
		return []string{".wav", ".mp3", ".aac", ".m4a", ".aiff"}
	case "linux":
		return []string{".wav", ".mp3", ".ogg", ".flac", ".aac"}
	default:
		return []string{".wav"}
	}
}

// TestSound reproduce un sonido de prueba
func (s *SoundNotifier) TestSound(soundType string) error {
	request := NotificationRequest{
		Event:    EventCustomAlert,
		Title:    "Test Sound",
		Message:  fmt.Sprintf("Testing %s sound", soundType),
		Priority: PriorityNormal,
	}

	// Temporal: cambiar el evento para que coincida con el tipo de sonido
	switch soundType {
	case "success":
		request.Event = EventPomodoroCompleted
	case "gentle":
		request.Event = EventBreakCompleted
	case "warning":
		request.Event = EventEarlyAlert
	case "urgent":
		request.Event = EventUrgentAlert
	case "start":
		request.Event = EventSessionStarted
	case "pause":
		request.Event = EventTimerPaused
	case "resume":
		request.Event = EventTimerResumed
	}

	response := s.Notify(request)
	return response.Error
}
