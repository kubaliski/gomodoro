# 🍅 Pomodoro Core

A robust, thread-safe, and UI-agnostic Pomodoro timer engine written in Go. This core library provides all the essential functionality for implementing the Pomodoro Technique in any type of application (CLI, web, desktop, mobile, etc.).

## ✨ Features

- **🔒 Thread-safe**: All operations are safe for concurrent use
- **🎯 UI-agnostic**: No UI dependencies, works with any interface
- **📊 Comprehensive stats**: Detailed session tracking and analytics
- **🔔 Event-driven**: Reactive architecture with pub-sub events
- **⚙️ Configurable**: Flexible configuration with validation
- **🧪 Testable**: Clean interfaces and dependency injection
- **📦 Zero dependencies**: Pure Go implementation

## 🚀 Quick Start

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
    // Create configuration
    cfg := config.DefaultConfig()

    // Create engine
    pomodoroEngine := engine.NewEngine(cfg)

    // Subscribe to events
    eventBus := pomodoroEngine.GetEventBus()
    eventBus.SubscribeFunc(events.PomodoroCompleted, func(event events.Event) {
        if data, ok := event.Data.(events.PomodoroEventData); ok {
            fmt.Printf("🍅 Pomodoro #%d completed!\n", data.Number)
        }
    })

    // Start the engine
    ctx := context.Background()
    pomodoroEngine.Start(ctx)

    // Engine runs in background, handling timers and emitting events
    // Your UI layer subscribes to events and sends commands
}
```

## 📋 Core Concepts

### Configuration

```go
cfg := &config.Config{
    WorkDuration:      25 * time.Minute,
    ShortBreak:        5 * time.Minute,
    LongBreak:         15 * time.Minute,
    LongBreakInterval: 4,
}

// Validate configuration
if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}
```

### Engine Control

```go
// Start the engine
err := engine.Start(context.Background())

// Control timer
engine.Pause()
engine.Resume()
engine.Skip()

// Stop engine
engine.Stop()
```

### Event System

```go
eventBus := engine.GetEventBus()

// Subscribe to specific events
eventBus.SubscribeFunc(events.TimerTick, handleTimerTick)
eventBus.SubscribeFunc(events.PomodoroCompleted, handlePomodoroCompleted)

// Subscribe to all events
eventBus.SubscribeGlobalFunc(handleAllEvents)
```

### Statistics

```go
stats := engine.GetStats()
snapshot := stats.GetSnapshot()

fmt.Printf("Completed: %d pomodoros\n", snapshot.PomodorosCompleted)
fmt.Printf("Current streak: %d\n", snapshot.CurrentStreak)
fmt.Printf("Work efficiency: %.1f%%\n", snapshot.WorkEfficiency)
```

## 📖 API Reference

### Engine Interface

```go
type EngineInterface interface {
    Start(ctx context.Context) error
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

### Event Types

| Event Type          | Description            | Data Type           |
| ------------------- | ---------------------- | ------------------- |
| `TimerStarted`      | Timer begins           | `TimerEventData`    |
| `TimerTick`         | Every second update    | `TimerEventData`    |
| `TimerPaused`       | Timer paused           | `TimerEventData`    |
| `TimerResumed`      | Timer resumed          | `TimerEventData`    |
| `TimerCompleted`    | Timer finished         | `TimerEventData`    |
| `TimerSkipped`      | Timer skipped          | `TimerEventData`    |
| `PomodoroStarted`   | Work session begins    | `PomodoroEventData` |
| `PomodoroCompleted` | Work session completed | `PomodoroEventData` |
| `PomodoroSkipped`   | Work session skipped   | `PomodoroEventData` |
| `BreakStarted`      | Break begins           | `BreakEventData`    |
| `BreakCompleted`    | Break completed        | `BreakEventData`    |
| `BreakSkipped`      | Break skipped          | `BreakEventData`    |
| `StatsUpdated`      | Statistics changed     | `StatsEventData`    |

### Event Data Structures

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
    // ... more fields
}
```

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────┐
│                 YOUR APP                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │     CLI     │ │     Web     │ │   Desktop   ││
│  │   Handler   │ │   Handler   │ │   Handler   ││
│  └─────────────┘ └─────────────┘ └─────────────┘│
└─────────────────┬───────────────────────────────┘
                  │ Events & Commands
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

## 🔧 Configuration Options

```go
type Config struct {
    WorkDuration      time.Duration  // Work session length
    ShortBreak        time.Duration  // Short break length
    LongBreak         time.Duration  // Long break length
    LongBreakInterval int           // Pomodoros before long break
}
```

**Validation Rules:**

- Work duration: 1 minute - 2 hours
- Short break: 1 minute - 30 minutes
- Long break: 5 minutes - 1 hour
- Long break interval: 2 - 10 pomodoros
- Long break must be longer than short break

## 📊 Statistics

The stats package provides comprehensive session tracking:

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

**Available Methods:**

- `GetSnapshot()` - Current stats snapshot
- `GetQuickStats()` - Formatted quick display
- `GetStatsDisplay()` - Full formatted display
- `ExportJSON()` - Export to JSON
- `Reset()` - Reset all statistics

## 🧪 Testing

Each package is designed to be easily testable:

```go
func TestEngine(t *testing.T) {
    cfg := config.DefaultConfig()
    cfg.WorkDuration = 1 * time.Second // Fast test

    engine := engine.NewEngine(cfg)

    // Test events
    var events []events.Event
    engine.GetEventBus().SubscribeGlobalFunc(func(e events.Event) {
        events = append(events, e)
    })

    ctx := context.Background()
    engine.Start(ctx)

    // Wait and verify
    time.Sleep(2 * time.Second)
    assert.Contains(t, eventTypes(events), events.PomodoroStarted)
}
```

## 🔄 Thread Safety

All public methods are thread-safe:

- Multiple goroutines can safely call engine methods
- Event handlers run in separate goroutines
- Internal state is protected with mutexes
- Context cancellation is handled properly

## 📝 Example Implementations

### CLI Handler

```go
type CLIHandler struct {
    engine engine.EngineInterface
    // UI specific fields
}

func (h *CLIHandler) Run() {
    // Subscribe to events
    eventBus := h.engine.GetEventBus()
    eventBus.SubscribeFunc(events.TimerTick, h.updateDisplay)

    // Start engine
    ctx := context.Background()
    h.engine.Start(ctx)

    // Handle user input
    h.handleInput()
}
```

### Web Handler

```go
type WebHandler struct {
    engine engine.EngineInterface
    clients map[string]*websocket.Conn
}

func (h *WebHandler) Run() {
    // Subscribe to events and broadcast to websocket clients
    eventBus := h.engine.GetEventBus()
    eventBus.SubscribeGlobalFunc(h.broadcastEvent)

    // Setup HTTP handlers
    http.HandleFunc("/ws", h.handleWebSocket)
    http.HandleFunc("/api/start", h.handleStart)
    // etc...
}
```

## 📦 Installation

```bash
go get github.com/kubaliski/pomodoro-core
```

## 📄 License

MIT License - see LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📚 Documentation

For more detailed documentation and examples, visit the [documentation site](https://github.com/kubaliski/pomodoro-core/docs).

---

Made with ❤️ for productivity enthusiasts
