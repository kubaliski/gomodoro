# 🍅 Pomodoro Core

Motor robusto, thread-safe y UI-agnóstico para temporizadores Pomodoro escrito en Go. Esta biblioteca central proporciona toda la funcionalidad esencial para implementar la Técnica Pomodoro en cualquier tipo de aplicación (CLI, web, escritorio, móvil, etc.).

## ✨ Características

- **🔒 Thread-safe**: Todas las operaciones son seguras para uso concurrente
- **🎯 UI-agnóstico**: Sin dependencias de UI, funciona con cualquier interfaz
- **📊 Estadísticas completas**: Seguimiento detallado de sesiones y análisis
- **🔔 Basado en eventos**: Arquitectura reactiva con eventos pub-sub
- **⚙️ Configurable**: Configuración flexible con validación
- **🧪 Testeable**: Interfaces limpias e inyección de dependencias
- **📦 Sin dependencias**: Implementación pura en Go

## 🚀 Inicio Rápido

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kubaliski/pomodoro-core/config"
    "github.com/kubaliski/pomodoro-core/engine"
    "github.com/kubaliski/pomodoro-core/events"
)

func main() {
    // Crear configuración
    cfg := config.DefaultConfig()

    // Crear motor
    pomodoroEngine := engine.NewEngine(cfg)

    // Suscribirse a eventos
    eventBus := pomodoroEngine.GetEventBus()
    eventBus.SubscribeFunc(events.PomodoroCompleted, func(event events.Event) {
        if data, ok := event.Data.(events.PomodoroEventData); ok {
            fmt.Printf("🍅 ¡Pomodoro #%d completado!\n", data.Number)
        }
    })

    // Iniciar el motor
    ctx := context.Background()
    pomodoroEngine.Start(ctx)
    pomodoroEngine.StartFirstSession()

    // El motor funciona en segundo plano, manejando timers y emitiendo eventos
    // Tu capa de UI se suscribe a eventos y envía comandos
}
```

## 📋 Conceptos Centrales

### Configuración

```go
cfg := &config.Config{
    WorkDuration:      25 * time.Minute,
    ShortBreak:        5 * time.Minute,
    LongBreak:         15 * time.Minute,
    LongBreakInterval: 4,
}

// Validar configuración
if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}
```

### Control del Motor

```go
// Iniciar el motor
err := engine.Start(context.Background())

// Controlar timer
engine.Pause()
engine.Resume()
engine.Skip()

// Detener motor
engine.Stop()
```

### Sistema de Eventos

```go
eventBus := engine.GetEventBus()

// Suscribirse a eventos específicos
eventBus.SubscribeFunc(events.TimerTick, manejarTickTimer)
eventBus.SubscribeFunc(events.PomodoroCompleted, manejarPomodoroCompletado)

// Suscribirse a todos los eventos
eventBus.SubscribeGlobalFunc(manejarTodosLosEventos)
```

### Estadísticas

```go
stats := engine.GetStats()
snapshot := stats.GetSnapshot()

fmt.Printf("Completados: %d pomodoros\n", snapshot.PomodorosCompleted)
fmt.Printf("Racha actual: %d\n", snapshot.CurrentStreak)
fmt.Printf("Eficiencia de trabajo: %.1f%%\n", snapshot.WorkEfficiency)
```

## 📖 Referencia de API

### Interfaz del Motor

```go
type EngineInterface interface {
    Start(ctx context.Context) error
    StartFirstSession() error
    Stop() error
    Pause() error
    Resume() error
    Skip() error
    GetState() State
    GetCurrentSession() SessionType
    GetPomodoroCount() int
    IsRunning() bool
    GetStats() *stats.SessionStats
    GetEventBus() *events.EventBus
    GetConfig() *config.Config
}
```

### Tipos de Eventos

| Tipo de Evento      | Descripción                  | Tipo de Datos       |
| ------------------- | ---------------------------- | ------------------- |
| `TimerStarted`      | Timer inicia                 | `TimerEventData`    |
| `TimerTick`         | Actualización cada segundo   | `TimerEventData`    |
| `TimerPaused`       | Timer pausado                | `TimerEventData`    |
| `TimerResumed`      | Timer reanudado              | `TimerEventData`    |
| `TimerCompleted`    | Timer terminado              | `TimerEventData`    |
| `TimerSkipped`      | Timer saltado                | `TimerEventData`    |
| `PomodoroStarted`   | Sesión de trabajo inicia     | `PomodoroEventData` |
| `PomodoroCompleted` | Sesión de trabajo completada | `PomodoroEventData` |
| `PomodoroSkipped`   | Sesión de trabajo saltada    | `PomodoroEventData` |
| `BreakStarted`      | Descanso inicia              | `BreakEventData`    |
| `BreakCompleted`    | Descanso completado          | `BreakEventData`    |
| `BreakSkipped`      | Descanso saltado             | `BreakEventData`    |
| `StatsUpdated`      | Estadísticas cambiaron       | `StatsEventData`    |

### Estructuras de Datos de Eventos

```go
type TimerEventData struct {
    Remaining    time.Duration
    Total        time.Duration
    State        string    // "TRABAJO", "DESCANSO", "DESCANSO LARGO"
    Status       string    // "RUNNING", "PAUSED", "STOPPED"
    Progress     float64   // 0.0 - 1.0
    SessionCount int
}

type PomodoroEventData struct {
    Number       int
    Duration     time.Duration
    ActualTime   time.Duration
    StartTime    time.Time
    EndTime      time.Time
}

type StatsEventData struct {
    PomodorosCompleted int
    CurrentStreak      int
    BestStreak         int
    TotalWorkTime      time.Duration
    WorkEfficiency     float64
    // ... más campos
}
```

## 🏗️ Arquitectura

```
┌─────────────────────────────────────────────────┐
│                 TU APLICACIÓN                   │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │     CLI     │ │   Discord   │ │     Web     ││
│  │   Handler   │ │   Handler   │ │   Handler   ││
│  └─────────────┘ └─────────────┘ └─────────────┘│
└─────────────────┬───────────────────────────────┘
                  │ Eventos y Comandos
┌─────────────────▼───────────────────────────────┐
│              POMODORO CORE                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │   Engine    │ │   Events    │ │    Stats    ││
│  └─────────────┘ └─────────────┘ └─────────────┘│
│  ┌─────────────┐ ┌─────────────┐                │
│  │   Timer     │ │   Config    │                │
│  └─────────────┘ └─────────────┘                │
└─────────────────────────────────────────────────┘
```

## 🔧 Opciones de Configuración

```go
type Config struct {
    WorkDuration      time.Duration  // Duración de sesión de trabajo
    ShortBreak        time.Duration  // Duración de descanso corto
    LongBreak         time.Duration  // Duración de descanso largo
    LongBreakInterval int           // Pomodoros antes del descanso largo
}
```

**Reglas de Validación:**

- Duración de trabajo: 1 minuto - 2 horas
- Descanso corto: 1 minuto - 30 minutos
- Descanso largo: 5 minutos - 1 hora
- Intervalo de descanso largo: 2 - 10 pomodoros
- El descanso largo debe ser mayor que el corto

## 📊 Estadísticas

El paquete stats proporciona seguimiento completo de sesiones:

```go
type StatsSnapshot struct {
    PomodorosCompleted  int
    PomodorosSkipped    int
    BreaksCompleted     int
    BreaksSkipped       int
    CurrentStreak       int
    BestStreak          int
    TotalWorkTime       time.Duration
    TotalBreakTime      time.Duration
    SessionDuration     time.Duration
    WorkEfficiency      float64
    TotalSessions       int
}
```

**Métodos Disponibles:**

- `GetSnapshot()` - Instantánea actual de estadísticas
- `GetQuickStats()` - Visualización rápida formateada
- `GetStatsDisplay()` - Visualización completa formateada
- `ExportJSON()` - Exportar a JSON
- `Reset()` - Reiniciar todas las estadísticas

## 🧪 Testing

Cada paquete está diseñado para ser fácilmente testeable:

```go
func TestEngine(t *testing.T) {
    cfg := config.DefaultConfig()
    cfg.WorkDuration = 1 * time.Second // Test rápido

    engine := engine.NewEngine(cfg)

    // Testear eventos
    var events []events.Event
    engine.GetEventBus().SubscribeGlobalFunc(func(e events.Event) {
        events = append(events, e)
    })

    ctx := context.Background()
    engine.Start(ctx)
    engine.StartFirstSession()

    // Esperar y verificar
    time.Sleep(2 * time.Second)
    assert.Contains(t, eventTypes(events), events.PomodoroStarted)
}
```

## 🔄 Thread Safety

Todos los métodos públicos son thread-safe:

- Múltiples goroutines pueden llamar métodos del motor de forma segura
- Los manejadores de eventos se ejecutan en goroutines separadas
- El estado interno está protegido con mutexes
- La cancelación de contexto se maneja correctamente

## 📝 Ejemplos de Implementación

### Manejador CLI

```go
type CLIHandler struct {
    engine engine.EngineInterface
    // campos específicos de UI
}

func (h *CLIHandler) Run() {
    // Suscribirse a eventos
    eventBus := h.engine.GetEventBus()
    eventBus.SubscribeFunc(events.TimerTick, h.actualizarPantalla)

    // Iniciar motor
    ctx := context.Background()
    h.engine.Start(ctx)

    // Manejar entrada del usuario
    h.manejarEntrada()
}
```

### Manejador Web

```go
type WebHandler struct {
    engine engine.EngineInterface
    clients map[string]*websocket.Conn
}

func (h *WebHandler) Run() {
    // Suscribirse a eventos y transmitir a clientes websocket
    eventBus := h.engine.GetEventBus()
    eventBus.SubscribeGlobalFunc(h.transmitirEvento)

    // Configurar manejadores HTTP
    http.HandleFunc("/ws", h.manejarWebSocket)
    http.HandleFunc("/api/start", h.manejarInicio)
    // etc...
}
```

## 📦 Instalación

```go
go get github.com/kubaliski/pomodoro-core
```

## 🤝 Contribuir

1. Fork el repositorio
2. Crea una rama de característica
3. Agrega tests para nueva funcionalidad
4. Asegúrate de que todos los tests pasen
5. Envía un pull request

## 📄 Licencia

Licencia MIT - ver archivo LICENSE para detalles.

## 📚 Documentación

Para documentación más detallada y ejemplos, visita el [sitio de documentación](https://github.com/kubaliski/pomodoro-core/docs).

---

Hecho con ❤️ para entusiastas de la productividad
