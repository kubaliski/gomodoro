package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/kubaliski/gomodoro/apps/discord/internal/bot"
	"github.com/kubaliski/gomodoro/apps/discord/internal/manager"
	"github.com/kubaliski/pomodoro-core/config"
)

func main() {
	// Intentar cargar archivo .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Cargar token del bot desde variable de entorno
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable is required")
	}

	// Configuración por defecto del pomodoro
	pomodoroConfig := config.DefaultConfig()

	// Crear el manager de sesiones
	sessionManager := manager.NewSessionManager(pomodoroConfig)

	// Crear y configurar el bot
	discordBot, err := bot.NewBot(token, sessionManager)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Crear contexto con cancelación
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Iniciar el bot
	if err := discordBot.Start(ctx); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}

	log.Println("🍅 Discord Pomodoro Bot is running. Press CTRL+C to exit.")

	// Esperar señal de interrupción
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down bot...")
	discordBot.Stop()
}
