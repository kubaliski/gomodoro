#!/bin/bash
# testing-scripts.sh
# Scripts para facilitar el testing del bot

echo "🧪 Pomodoro Bot Testing Scripts"
echo "================================"

# Función para mostrar logs en tiempo real con filtros
monitor_logs() {
    echo "📊 Monitoring logs for testing..."
    echo "Ctrl+C to stop monitoring"

    # Mostrar logs en tiempo real con colores
    tail -f logs/bot.log | while read line; do
        if [[ $line == *"❌"* ]]; then
            echo -e "\033[31m$line\033[0m"  # Rojo para errores
        elif [[ $line == *"✅"* ]]; then
            echo -e "\033[32m$line\033[0m"  # Verde para éxito
        elif [[ $line == *"📱"* ]]; then
            echo -e "\033[34m$line\033[0m"  # Azul para DMs
        elif [[ $line == *"📢"* ]]; then
            echo -e "\033[33m$line\033[0m"  # Amarillo para fallbacks
        else
            echo "$line"
        fi
    done
}

# Función para analizar logs después del testing
analyze_logs() {
    echo "📈 Analyzing test results..."

    if [ ! -f "logs/bot.log" ]; then
        echo "❌ No log file found. Make sure bot is running and logging to logs/bot.log"
        return
    fi

    echo ""
    echo "=== TEST RESULTS SUMMARY ==="

    # Contar eventos exitosos
    dm_success=$(grep -c "📱 DM sent successfully" logs/bot.log)
    fallback_count=$(grep -c "📢.*fallback" logs/bot.log)
    welcome_sent=$(grep -c "👋 Welcome message sent" logs/bot.log)
    sessions_started=$(grep -c "🚀 Starting new session" logs/bot.log)
    errors=$(grep -c "❌" logs/bot.log)

    echo "📊 Sessions Started: $sessions_started"
    echo "📱 DM Notifications Sent: $dm_success"
    echo "📢 Fallback Notifications: $fallback_count"
    echo "👋 Welcome Messages: $welcome_sent"
    echo "❌ Errors: $errors"

    # Calcular métricas
    total_notifications=$((dm_success + fallback_count))
    if [ $total_notifications -gt 0 ]; then
        dm_success_rate=$((dm_success * 100 / total_notifications))
        echo "📈 DM Success Rate: ${dm_success_rate}%"
    fi

    echo ""
    echo "=== RECENT ERRORS ==="
    if [ $errors -gt 0 ]; then
        tail -n 50 logs/bot.log | grep "❌" | tail -n 5
    else
        echo "✅ No errors found!"
    fi

    echo ""
    echo "=== RECENT FALLBACKS ==="
    if [ $fallback_count -gt 0 ]; then
        tail -n 50 logs/bot.log | grep "📢.*fallback" | tail -n 3
    else
        echo "✅ No fallbacks needed (all DMs successful)"
    fi
}

# Función para testing rápido de comandos
test_commands() {
    echo "🎯 Quick Command Testing Guide"
    echo "=============================="
    echo ""
    echo "Test these commands in Discord:"
    echo ""
    echo "1. BASIC FLOW:"
    echo "   /pomodoro"
    echo "   (Wait for welcome DM and first notification)"
    echo ""
    echo "2. CONTROL COMMANDS:"
    echo "   /pomodoro-status"
    echo "   /pomodoro-pause"
    echo "   /pomodoro-resume"
    echo "   /pomodoro-skip"
    echo "   /pomodoro-stats"
    echo "   /pomodoro-stop"
    echo ""
    echo "3. CUSTOM CONFIG:"
    echo "   /pomodoro work:15 short_break:3 long_break:10"
    echo ""
    echo "4. EDGE CASES:"
    echo "   - Try /pomodoro when session already active"
    echo "   - Use commands without active session"
    echo "   - Test with DMs disabled"
    echo ""
    echo "Monitor results with: ./testing-scripts.sh monitor"
}

# Función para cargar variables de entorno desde .env
load_env() {
    if [ -f ".env" ]; then
        echo "📂 Loading environment from .env file..."
        export $(grep -v '^#' .env | xargs)
        return 0
    else
        echo "⚠️  No .env file found in current directory"
        return 1
    fi
}

# Función para verificar setup
check_setup() {
    echo "🔧 Checking testing setup..."
    local os=$(detect_os)
    echo "🖥️  Detected OS: $os"

    # Cargar .env si existe
    load_env

    # Verificar bot executable
    local bot_exists=false
    if [ -f "./pomodoro-bot" ] || [ -f "./pomodoro-bot.exe" ]; then
        bot_exists=true
    fi

    if [ "$bot_exists" = false ]; then
        echo "❌ Bot executable not found. Building it now..."
        go build -o pomodoro-bot main.go
        if [ $? -ne 0 ]; then
            echo "❌ Failed to build bot. Make sure you're in the correct directory and Go is installed."
            return 1
        fi
        echo "✅ Bot built successfully"
    fi

    # Verificar variables de entorno
    if [ -z "$DISCORD_BOT_TOKEN" ]; then
        echo "❌ DISCORD_BOT_TOKEN not set"
        echo "💡 Make sure you have either:"
        echo "   - A .env file with DISCORD_BOT_TOKEN=your_token"
        echo "   - Export DISCORD_BOT_TOKEN environment variable"
        if [ -f ".env" ]; then
            echo "📂 Found .env file, but DISCORD_BOT_TOKEN may be missing or commented out"
        fi
        return 1
    fi

    # Crear directorio de logs si no existe
    mkdir -p logs

    echo "✅ Setup looks good!"
    echo "🤖 Bot token: ${DISCORD_BOT_TOKEN:0:20}..."
    echo ""
    echo "To start testing:"
    echo "1. ./pomodoro-bot > logs/bot.log 2>&1 &"
    echo "2. ./testing-scripts.sh monitor (in another terminal)"
    echo "3. Test commands in Discord"
    echo "4. ./testing-scripts.sh analyze"
}

# Detectar el sistema operativo
detect_os() {
    if [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ -n "$WINDIR" ]]; then
        echo "windows"
    else
        echo "unix"
    fi
}

# Función para verificar si el bot está corriendo (compatible con Windows)
is_bot_running() {
    local os=$(detect_os)
    if [ "$os" = "windows" ]; then
        # En Windows, usar tasklist
        tasklist.exe 2>/dev/null | grep -q "pomodoro-bot.exe"
    else
        # En Unix, usar pgrep
        pgrep -f "pomodoro-bot" > /dev/null
    fi
}

# Función para matar el proceso del bot (compatible con Windows)
kill_bot_process() {
    local os=$(detect_os)
    if [ "$os" = "windows" ]; then
        # En Windows, usar taskkill
        taskkill.exe //F //IM pomodoro-bot.exe 2>/dev/null
    else
        # En Unix, usar pkill
        pkill -f "pomodoro-bot"
    fi
}

# Función para iniciar bot con variables de entorno
start_bot() {
    echo "🚀 Starting bot for testing..."
    local os=$(detect_os)
    echo "🖥️  Detected OS: $os"

    # Cargar .env
    load_env

    if [ -z "$DISCORD_BOT_TOKEN" ]; then
        echo "❌ Cannot start bot: DISCORD_BOT_TOKEN not set"
        return 1
    fi

    # Verificar si el bot ya está corriendo
    if is_bot_running; then
        echo "⚠️  Bot is already running. Stop it first with: ./testing-scripts.sh stop"
        return 1
    fi

    # Crear directorio de logs
    mkdir -p logs

    # Compilar el bot si no existe
    if [ ! -f "./pomodoro-bot" ] && [ ! -f "./pomodoro-bot.exe" ]; then
        echo "🔨 Building bot..."
        go build -o pomodoro-bot main.go
        if [ $? -ne 0 ]; then
            echo "❌ Failed to build bot"
            return 1
        fi
    fi

    # Determinar el ejecutable correcto
    local bot_exe="./pomodoro-bot"
    if [ "$os" = "windows" ] && [ -f "./pomodoro-bot.exe" ]; then
        bot_exe="./pomodoro-bot.exe"
    fi

    # Iniciar bot
    echo "🤖 Starting bot with logging..."
    $bot_exe > logs/bot.log 2>&1 &
    BOT_PID=$!

    # Guardar PID para poder parar el bot después
    echo $BOT_PID > logs/bot.pid

    sleep 3

    # Verificar que el bot inició correctamente
    if is_bot_running; then
        echo "✅ Bot started successfully (PID: $BOT_PID)"
        echo "📊 Monitor logs with: ./testing-scripts.sh monitor"
        echo "🛑 Stop bot with: ./testing-scripts.sh stop"

        # Mostrar las primeras líneas del log para verificar
        echo ""
        echo "📝 First few log lines:"
        head -n 5 logs/bot.log 2>/dev/null || echo "⚠️  Log file not ready yet, wait a moment"
    else
        echo "❌ Bot failed to start. Check logs:"
        if [ -f "logs/bot.log" ]; then
            tail -n 10 logs/bot.log
        else
            echo "No log file found"
        fi
        return 1
    fi
}

# Función para parar el bot
stop_bot() {
    echo "🛑 Stopping bot..."
    local os=$(detect_os)

    if [ -f "logs/bot.pid" ]; then
        BOT_PID=$(cat logs/bot.pid)
        if [ "$os" = "windows" ]; then
            # En Windows, usar taskkill con PID
            taskkill.exe //F //PID $BOT_PID 2>/dev/null
            if [ $? -eq 0 ]; then
                echo "✅ Bot stopped (PID: $BOT_PID)"
            else
                echo "⚠️  Bot PID not found, trying to kill by name..."
                kill_bot_process
            fi
        else
            # En Unix, usar kill
            if kill $BOT_PID 2>/dev/null; then
                echo "✅ Bot stopped (PID: $BOT_PID)"
            else
                echo "⚠️  Bot PID not found, trying to kill by name..."
                kill_bot_process
            fi
        fi
        rm -f logs/bot.pid
    else
        echo "⚠️  No PID file found, trying to kill by name..."
        if kill_bot_process; then
            echo "✅ Bot stopped"
        else
            echo "❌ No bot process found"
        fi
    fi
}

# Función para limpiar logs de testing
clean_logs() {
    echo "🧹 Cleaning test logs..."
    rm -f logs/bot.log
    echo "✅ Logs cleaned. Ready for fresh testing."
}

# Función principal
case "$1" in
    "monitor")
        monitor_logs
        ;;
    "analyze")
        analyze_logs
        ;;
    "commands")
        test_commands
        ;;
    "setup")
        check_setup
        ;;
    "start")
        start_bot
        ;;
    "stop")
        stop_bot
        ;;
    "restart")
        stop_bot
        sleep 1
        start_bot
        ;;
    "clean")
        clean_logs
        ;;
    *)
        echo "Usage: $0 {setup|start|stop|restart|monitor|commands|analyze|clean}"
        echo ""
        echo "Commands:"
        echo "  setup    - Check if everything is ready for testing"
        echo "  start    - Start the bot with logging (reads .env automatically)"
        echo "  stop     - Stop the running bot"
        echo "  restart  - Stop and start the bot"
        echo "  monitor  - Monitor logs in real-time during testing"
        echo "  commands - Show testing command guide"
        echo "  analyze  - Analyze test results from logs"
        echo "  clean    - Clean logs for fresh testing"
        echo ""
        echo "Quick testing workflow:"
        echo "1. ./testing-scripts.sh setup"
        echo "2. ./testing-scripts.sh start"
        echo "3. ./testing-scripts.sh monitor (in another terminal)"
        echo "4. Use Discord to test commands"
        echo "5. ./testing-scripts.sh analyze"
        echo "6. ./testing-scripts.sh stop"
        ;;
esac