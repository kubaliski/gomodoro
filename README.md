# Gomodoro - Pomodoro Timer Suite

Un ecosistema completo de aplicaciones Pomodoro desarrollado en Go, que incluye una biblioteca central robusta y múltiples interfaces de usuario.

## 🌟 Características Principales

- **🏗️ Arquitectura modular**: Core library + aplicaciones específicas
- **🔒 Thread-safe**: Operaciones seguras para uso concurrente
- **🎯 UI-agnóstico**: El core funciona con cualquier interfaz
- **📊 Estadísticas completas**: Seguimiento detallado de sesiones y productividad
- **🔔 Sistema de eventos**: Arquitectura reactiva con pub-sub
- **⚙️ Altamente configurable**: Configuración flexible con validación
- **🧪 Testeable**: Interfaces limpias e inyección de dependencias

## 📦 Estructura del Proyecto

```
gomodoro/
├── 📚 core/              # Biblioteca central del motor Pomodoro
├── 🖥️ apps/
│   ├── cli/              # Aplicación de línea de comandos
│   └── 💬 discord/       # Bot de Discord
├── 📄 README.md          # Este archivo
├── 🔧 go.work           # Go workspace para desarrollo
└── 🚫 .gitignore        # Archivos ignorados por Git
```

## 🚀 Aplicaciones Disponibles

### 🖥️ CLI - Línea de Comandos

Temporizador Pomodoro completo para la terminal con interfaz colorida e interactiva.

**Características:**

- ⏱️ Timer configurable (25/5/15 minutos por defecto)
- ⏸️ Pausa, reanuda y salta sesiones
- 🔄 Ciclos automáticos trabajo/descanso
- 📊 Estadísticas detalladas en tiempo real
- 🎨 Interfaz colorida en terminal
- 🔔 Notificaciones del sistema

**Uso rápido:**

```bash
cd apps/cli
go run main.go
```

### 💬 Discord Bot

Bot completo para servidores de Discord con soporte multi-usuario.

**Características:**

- 👥 Sesiones simultáneas para múltiples usuarios
- 🎛️ Configuración personalizable por sesión
- 📱 Comandos slash modernos de Discord
- 🎨 Embeds ricos con colores y emojis
- 🔔 Notificaciones automáticas y menciones
- 📈 Estadísticas visuales con barras de progreso

**Comandos disponibles:**

- `/pomodoro` - Iniciar sesión
- `/pomodoro-pause` - Pausar
- `/pomodoro-resume` - Reanudar
- `/pomodoro-status` - Ver estado
- `/pomodoro-stats` - Estadísticas

## 🏗️ Core Library

El corazón del sistema - una biblioteca thread-safe que proporciona toda la funcionalidad del Pomodoro:

### Módulos Principales

- **🔧 Engine**: Motor principal que coordina timers y eventos
- **⏰ Timer**: Temporizador thread-safe con estado inmutable
- **📊 Stats**: Sistema de estadísticas con seguimiento completo
- **🔔 Events**: Bus de eventos para arquitectura reactiva
- **⚙️ Config**: Configuración validada y flexible

### Ejemplo de Uso

```go
package main

import (
    "context"
    "fmt"
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
        fmt.Println("🍅 ¡Pomodoro completado!")
    })

    // Iniciar motor
    ctx := context.Background()
    pomodoroEngine.Start(ctx)
    pomodoroEngine.StartFirstSession()

    // El motor funciona en segundo plano emitiendo eventos
}
```

## 🛠️ Desarrollo

### Requisitos

- **Go 1.24+**
- **Git**

### Configuración del Entorno

```bash
# Clonar el repositorio
git clone https://github.com/tu-usuario/gomodoro.git
cd gomodoro

# El proyecto usa Go Workspaces para desarrollo
go work sync

# Ejecutar tests
go test ./...
```

### Desarrollo de Aplicaciones

#### CLI

```bash
cd apps/cli
go run main.go
```

#### Discord Bot

```bash
cd apps/discord

# Configurar token (ver apps/discord/README.md)
cp .env.example .env
# Editar .env con tu token de Discord

go run main.go
```

### Creando una Nueva Aplicación

El core está diseñado para ser UI-agnóstico. Para crear una nueva aplicación:

1. **Crea tu directorio**: `apps/tu-app/`
2. **Inicializa go module**: `go mod init github.com/tu-usuario/gomodoro/apps/tu-app`
3. **Agrega dependencia del core**:
   ```go
   require github.com/kubaliski/pomodoro-core v0.0.0
   replace github.com/kubaliski/pomodoro-core => ../../core
   ```
4. **Implementa los event handlers** para tu UI específica
5. **Suscríbete a eventos** del motor

## 🧪 Testing

```bash
# Tests del core
cd core && go test ./...

# Tests de aplicaciones específicas
cd apps/cli && go test ./...
cd apps/discord && go test ./...

# Tests de todo el proyecto
go test ./...
```

## 📊 Estadísticas y Métricas

El sistema rastrea automáticamente:

- 🍅 **Pomodoros**: Completados, saltados, racha actual/mejor
- ☕ **Descansos**: Cortos, largos, completados, saltados
- ⏱️ **Tiempo**: Trabajo total, descanso total, duración de sesión
- 📈 **Eficiencia**: Porcentaje de pomodoros completados vs saltados
- 📋 **Sesiones**: Historial completo con timestamps

## ⚙️ Configuración

```go
type Config struct {
    WorkDuration      time.Duration  // Duración del trabajo (1min - 2h)
    ShortBreak        time.Duration  // Descanso corto (1min - 30min)
    LongBreak         time.Duration  // Descanso largo (5min - 1h)
    LongBreakInterval int           // Pomodoros antes del descanso largo (2-10)
}
```

**Configuración por defecto:**

- 🍅 Trabajo: 25 minutos
- ☕ Descanso corto: 5 minutos
- 🏖️ Descanso largo: 15 minutos
- 🔄 Intervalo: Cada 4 pomodoros

## 🔄 Eventos del Sistema

El core emite eventos que las aplicaciones pueden escuchar:

| Evento              | Descripción               | Cuándo se emite                    |
| ------------------- | ------------------------- | ---------------------------------- |
| `PomodoroStarted`   | Pomodoro iniciado         | Al comenzar período de trabajo     |
| `PomodoroCompleted` | Pomodoro completado       | Al terminar trabajo exitosamente   |
| `BreakStarted`      | Descanso iniciado         | Al comenzar descanso (corto/largo) |
| `BreakCompleted`    | Descanso completado       | Al terminar descanso               |
| `TimerTick`         | Tick del timer            | Cada segundo durante ejecución     |
| `TimerPaused`       | Timer pausado             | Al pausar sesión                   |
| `StatsUpdated`      | Estadísticas actualizadas | Al completar/saltar sesiones       |

## 🤝 Contribuir

1. **Fork** el repositorio
2. **Crea una rama** para tu feature (`git checkout -b feature/nueva-caracteristica`)
3. **Commit** tus cambios (`git commit -am 'Agrega nueva característica'`)
4. **Push** a la rama (`git push origin feature/nueva-caracteristica`)
5. **Abre un Pull Request**

### Convenciones de Código

- Usar `gofmt` para formatear código
- Agregar tests para nueva funcionalidad
- Documentar funciones públicas
- Seguir [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## 📄 Licencia

MIT License - ver archivo [LICENSE](LICENSE) para detalles.

## 🙏 Agradecimientos

- **Francesco Cirillo** por la [Técnica Pomodoro](https://francescocirillo.com/pages/pomodoro-technique)
- **Go Team** por el excelente lenguaje
- **Discord** por su API robusta
- **Comunidad Go** por las librerías y herramientas

---

Hecho con ❤️ para mejorar la productividad

## 📚 Enlaces Útiles

- [Documentación del Core](core/README.md)
- [Guía de la CLI](apps/cli/README.md)
- [Setup del Discord Bot](apps/discord/README.md)
- [Técnica Pomodoro Original](https://francescocirillo.com/pages/pomodoro-technique)
- [Go Workspaces](https://go.dev/ref/mod#workspaces)
