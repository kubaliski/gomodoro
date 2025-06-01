# 🖥️ Pomodoro CLI

Temporizador Pomodoro completo para la línea de comandos, desarrollado en Go. Una aplicación robusta que implementa la Técnica Pomodoro con interfaz colorida, estadísticas detalladas y notificaciones del sistema.

## ✨ Características

- ⏱️ **Timer configurable**: 25/5/15 minutos por defecto, personalizable
- ⏸️ **Control completo**: Pausa, reanuda y salta sesiones
- 🔄 **Ciclos automáticos**: Trabajo → Descanso Corto → Trabajo → Descanso Largo
- 📊 **Estadísticas en tiempo real**: Pomodoros completados, rachas, eficiencia
- 🎨 **Interfaz colorida**: Colores dinámicos según el estado del timer
- 🔔 **Notificaciones**: Alertas del sistema cuando cambian las sesiones
- 💾 **Persistencia**: Las estadísticas se mantienen durante la sesión
- 🎯 **Fácil de usar**: Controles intuitivos con teclas simples

## 🚀 Instalación y Uso

### Instalación

```bash
# Desde el directorio del proyecto
cd apps/cli
go build -o pomodoro main.go

# O ejecutar directamente
go run main.go
```

### Uso Básico

```bash
# Iniciar la aplicación
./pomodoro
# o
go run main.go
```

### Controles Interactivos

Una vez iniciada la aplicación, usa estas teclas:

| Tecla     | Acción                         |
| --------- | ------------------------------ |
| `ESPACIO` | Iniciar/Pausar timer           |
| `r`       | Reanudar timer pausado         |
| `s`       | Saltar sesión actual           |
| `q`       | Salir de la aplicación         |
| `h`       | Mostrar ayuda                  |
| `t`       | Alternar vista de estadísticas |

## 🎨 Interfaz de Usuario

### Vista Principal del Timer

```
🍅 POMODORO - SESIÓN DE TRABAJO

    ███████████████████████████████████████████████████
    ████████████████████████████████████░░░░░░░░░░░░░░░  75%
    ███████████████████████████████████████████████████

                         18:45
                    Tiempo restante

         [ESPACIO] Pausar  [s] Saltar  [q] Salir
```

### Vista de Estadísticas

```
📊 ESTADÍSTICAS DE LA SESIÓN

🍅 Pomodoros: 3 completados, 0 saltados
☕ Descansos: 2 completados, 0 saltados
🔥 Racha actual: 3 pomodoros
⭐ Mejor racha: 5 pomodoros
⏱️  Tiempo trabajado: 1h 15m
📈 Eficiencia: 100%

         [t] Ocultar estadísticas  [q] Salir
```

### Estados Visuales

- **🍅 Trabajo**: Barra roja/naranja, enfoque en productividad
- **☕ Descanso Corto**: Barra azul/cyan, relajación ligera
- **🏖️ Descanso Largo**: Barra verde/magenta, descanso profundo
- **⏸️ Pausado**: Barra amarilla/gris, estado de espera

## ⚙️ Configuración

### Configuración por Defecto

```go
// Duración estándar de la Técnica Pomodoro
Trabajo:         25 minutos
Descanso Corto:   5 minutos
Descanso Largo:  15 minutos
Intervalo:        4 pomodoros (antes del descanso largo)
```

### Configuración Personalizada

La configuración se puede modificar editando el archivo de configuración o a través de variables de entorno:

```bash
# Variables de entorno
export POMODORO_WORK_DURATION=30m
export POMODORO_SHORT_BREAK=10m
export POMODORO_LONG_BREAK=20m
export POMODORO_LONG_BREAK_INTERVAL=3
```

## 🔔 Notificaciones del Sistema

La aplicación envía notificaciones del sistema en momentos clave:

- 🍅 **Inicio de Pomodoro**: "¡Hora de concentrarse! Pomodoro iniciado"
- ✅ **Pomodoro Completado**: "¡Excelente! Pomodoro completado. Hora del descanso"
- ☕ **Inicio de Descanso**: "Tiempo de relajarse. Descanso iniciado"
- 🔄 **Fin de Descanso**: "¡De vuelta al trabajo! Descanso terminado"

## 📊 Sistema de Estadísticas

### Métricas Rastreadas

- **Pomodoros Completados/Saltados**: Conteo de sesiones de trabajo
- **Descansos Completados/Saltados**: Conteo de sesiones de descanso
- **Rachas**: Pomodoros consecutivos completados
- **Tiempo Total**: Trabajo y descanso acumulado
- **Eficiencia**: Porcentaje de pomodoros completados vs saltados
- **Duración de Sesión**: Tiempo total desde que inició la aplicación

### Visualización

Las estadísticas se muestran tanto en la interfaz principal como en una vista dedicada que puedes alternar con la tecla `t`.

## 🏗️ Arquitectura del Proyecto

```
apps/cli/
├── main.go                     # Punto de entrada principal
├── internal/
│   ├── handlers/
│   │   ├── cli_handler.go      # Manejador principal de la CLI
│   │   ├── command_processor.go # Procesamiento de comandos
│   │   ├── event_handlers.go   # Manejadores de eventos del core
│   │   ├── input_manager.go    # Gestión de entrada del usuario
│   │   └── ui_helpers.go       # Utilidades de interfaz
│   ├── notifications/
│   │   ├── manager.go          # Gestión de notificaciones
│   │   ├── config.go           # Configuración de notificaciones
│   │   └── sound.go            # Sonidos (futuro)
│   └── ui/
│       ├── display.go          # Renderizado de interfaz
│       ├── colors.go           # Esquemas de colores
│       └── stats_display.go    # Visualización de estadísticas
├── go.mod                      # Dependencias del módulo
└── README.md                   # Este archivo
```

## 🔧 Desarrollo

### Requisitos

- Go 1.21+
- Terminal compatible con colores ANSI
- Sistema operativo: Windows, macOS, Linux

### Compilación

```bash
# Compilación simple
go build -o pomodoro main.go

# Compilación con optimizaciones
go build -ldflags="-s -w" -o pomodoro main.go

# Cross-compilation para diferentes plataformas
GOOS=windows GOARCH=amd64 go build -o pomodoro.exe main.go
GOOS=linux GOARCH=amd64 go build -o pomodoro-linux main.go
GOOS=darwin GOARCH=amd64 go build -o pomodoro-mac main.go
```

### Testing

```bash
# Ejecutar tests
go test ./...

# Tests con cobertura
go test -cover ./...

# Tests verbosos
go test -v ./...
```

### Debugging

```bash
# Ejecutar con logs de debug
go run main.go -debug

# Usar delve para debugging
dlv debug main.go
```

## 🎯 Casos de Uso

### Desarrollador/Programador

```bash
# Sesiones de codificación enfocada
./pomodoro
# Usar para sprints de desarrollo, debugging, code reviews
```

### Estudiante

```bash
# Sesiones de estudio concentrado
./pomodoro
# Ideal para lectura, escritura, resolución de problemas
```

### Trabajador Remoto

```bash
# Gestión de tiempo en trabajo desde casa
./pomodoro
# Mantener productividad y evitar distracciones
```

### Escritor/Creativo

```bash
# Bloques de escritura o trabajo creativo
./pomodoro
# Mantener el flujo creativo con descansos regulares
```

## 🔧 Personalización

### Modificando Colores

Los colores se pueden personalizar editando `internal/ui/colors.go`:

```go
// Esquemas de colores personalizables
var ColorSchemes = map[string]ColorScheme{
    "work": {
        Primary:   color.New(color.FgRed, color.Bold),
        Secondary: color.New(color.FgHiRed),
        Background: color.New(color.BgRed, color.FgWhite),
    },
    "break": {
        Primary:   color.New(color.FgCyan, color.Bold),
        Secondary: color.New(color.FgHiCyan),
        Background: color.New(color.BgCyan, color.FgBlack),
    },
}
```

### Agregando Nuevos Comandos

Para agregar nuevos comandos de teclado:

1. Modifica `internal/handlers/input_manager.go`
2. Agrega el manejador en `internal/handlers/command_processor.go`
3. Actualiza la ayuda en `internal/ui/display.go`

## 🚀 Características Futuras

### En Desarrollo

- [ ] 🔊 **Sonidos personalizables**: Alertas auditivas configurables
- [ ] 📁 **Configuración por archivo**: Archivos de configuración YAML/JSON
- [ ] 📈 **Estadísticas históricas**: Persistencia entre sesiones
- [ ] 🎨 **Temas personalizables**: Múltiples esquemas de colores
- [ ] 📱 **Integración con móvil**: Notificaciones cross-platform

### Planificadas

- [ ] 🌐 **Modo servidor**: API REST para integración web
- [ ] 📊 **Exportación de datos**: CSV, JSON para análisis externo
- [ ] 🔗 **Integración con herramientas**: Slack, Discord, Todoist
- [ ] 📅 **Programación**: Sesiones automáticas por horario
- [ ] 🏆 **Logros/Achievements**: Gamificación de productividad

## 🐛 Troubleshooting

### Problemas Comunes

**Los colores no se muestran correctamente:**

```bash
# Verificar soporte ANSI del terminal
echo $TERM

# En Windows, usar Windows Terminal o PowerShell 7+
# En macOS/Linux, la mayoría de terminals modernos funcionan
```

**La aplicación no responde a teclas:**

```bash
# Asegurar que el terminal tiene focus
# Verificar que no hay otros procesos interceptando input
# Reiniciar la aplicación si es necesario
```

**Notificaciones no aparecen:**

```bash
# En Linux: verificar que libnotify está instalado
sudo apt-get install libnotify-bin

# En macOS: permisos de notificaciones en Preferencias del Sistema
# En Windows: verificar configuración de notificaciones
```

### Logs de Debug

```bash
# Ejecutar con información de debug
DEBUG=true go run main.go

# Ver logs detallados del core
CORE_DEBUG=true go run main.go
```

## 🤝 Contribuir

### Cómo Contribuir

1. **Fork** el repositorio
2. **Crea una rama** para tu feature (`git checkout -b feature/nueva-caracteristica`)
3. **Implement** la funcionalidad con tests
4. **Asegúrate** de que todos los tests pasen
5. **Actualiza** la documentación si es necesario
6. **Commit** tus cambios (`git commit -am 'Agrega nueva característica'`)
7. **Push** a tu rama (`git push origin feature/nueva-caracteristica`)
8. **Abre un Pull Request**

### Convenciones de Código

- Usar `gofmt` para formatear código
- Seguir [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Agregar tests para nueva funcionalidad
- Documentar funciones públicas
- Mantener cobertura de tests > 80%

### Áreas de Contribución

- 🐛 **Bug fixes**: Corregir problemas reportados
- ✨ **Nuevas características**: Implementar funcionalidades planificadas
- 📚 **Documentación**: Mejorar READMEs y documentación de código
- 🧪 **Tests**: Aumentar cobertura de tests
- 🎨 **UI/UX**: Mejorar interfaz y experiencia de usuario
- 🔧 **Performance**: Optimizaciones de rendimiento

## 📄 Licencia

MIT License - ver archivo [LICENSE](../../LICENSE) para detalles.

## 🙏 Agradecimientos

- **Francesco Cirillo** por la [Técnica Pomodoro](https://francescocirillo.com/pages/pomodoro-technique) original
- **Go Community** por las excelentes librerías de terminal
- **Contributors** que han mejorado esta aplicación

## 📚 Enlaces Útiles

- [🏠 Proyecto Principal](../../README.md)
- [🔧 Documentación del Core](../../core/README.md)
- [💬 Discord Bot](../discord/README.md)
- [📖 Técnica Pomodoro Original](https://francescocirillo.com/pages/pomodoro-technique)
- [🐛 Reportar Issues](https://github.com/tu-usuario/gomodoro/issues)

---

Hecho con ❤️ para maximizar tu productividad desde la terminal
