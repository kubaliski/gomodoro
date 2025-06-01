# 💬 Gomodoro Discord Bot

Bot de Discord que lleva la Técnica Pomodoro a tu servidor usando el motor central de Gomodoro. Diseñado para equipos y comunidades que quieren mejorar su productividad de manera colaborativa.

## ✨ Características

- **👥 Sesiones simultáneas**: Cada usuario puede tener su propia sesión independiente
- **🎛️ Timers personalizables**: Configura duración de trabajo y descansos por sesión
- **🎨 Integración rica con Discord**: Embeds hermosos, comandos slash y notificaciones
- **📊 Seguimiento de estadísticas**: Rastrea tu productividad a través de sesiones
- **⏸️ Control completo**: Pausa, reanuda y salta sesiones
- **🔔 Notificaciones inteligentes**: Avisos automáticos y menciones personales
- **📱 Comandos slash modernos**: Interfaz nativa de Discord
- **🔒 Thread-safe**: Manejo seguro de múltiples usuarios simultáneos

## 🚀 Configuración Rápida

### 1. Crear Aplicación de Discord

1. Ve al [Portal de Desarrolladores de Discord](https://discord.com/developers/applications)
2. Crea una nueva aplicación
3. Ve a la sección "Bot" y crea un bot
4. Copia el token del bot
5. Ve a "OAuth2" > "URL Generator"
6. Selecciona scopes: `bot` y `applications.commands`
7. Selecciona permisos: `Send Messages`, `Use Slash Commands`, `Embed Links`

### 2. Configuración del Entorno

```bash
# Ir al directorio del bot
cd apps/discord

# Copiar archivo de ejemplo
cp .env.example .env

# Editar .env con tu token
# DISCORD_BOT_TOKEN=tu_token_aquí
# DISCORD_APPLICATION_ID=tu_application_id_aquí
```

### 3. Instalar Dependencias

```bash
go mod tidy
```

### 4. Ejecutar el Bot

```bash
go run main.go
```

### 5. Invitar el Bot a tu Servidor

Usa la URL generada en el paso 1, o construye manualmente:

```
https://discord.com/api/oauth2/authorize?client_id=TU_APPLICATION_ID&permissions=2147695616&scope=bot%20applications.commands
```

## 🎮 Comandos Disponibles

| Comando            | Descripción                               | Opciones                                      |
| ------------------ | ----------------------------------------- | --------------------------------------------- |
| `/pomodoro`        | Iniciar una nueva sesión de pomodoro      | `work`, `short_break`, `long_break` (minutos) |
| `/pomodoro-stop`   | Detener tu sesión actual                  | -                                             |
| `/pomodoro-pause`  | Pausar tu sesión actual                   | -                                             |
| `/pomodoro-resume` | Reanudar tu sesión pausada                | -                                             |
| `/pomodoro-skip`   | Saltar el pomodoro o descanso actual      | -                                             |
| `/pomodoro-status` | Verificar el estado actual de tu pomodoro | -                                             |
| `/pomodoro-stats`  | Ver tus estadísticas de pomodoro          | -                                             |

## 📱 Ejemplos de Uso

### Sesión Básica

```
/pomodoro
```

Inicia un pomodoro con configuración por defecto (25min trabajo, 5min descanso corto, 15min descanso largo)

### Sesión Personalizada

```
/pomodoro work:30 short_break:10 long_break:20
```

Trabajo de 30 minutos, descanso corto de 10 minutos, descanso largo de 20 minutos

### Control de Sesión

```
/pomodoro-pause     # Pausar sesión actual
/pomodoro-resume    # Reanudar sesión pausada
/pomodoro-skip      # Saltar al siguiente período
```

### Monitoreo

```
/pomodoro-status    # Ver estado actual
/pomodoro-stats     # Ver estadísticas detalladas
```

## 🔔 Sistema de Notificaciones

### Notificaciones Automáticas

El bot envía notificaciones automáticas en momentos clave:

#### Durante el Trabajo (25 min por defecto)

- **10 minutos restantes**: "⏰ Quedan 10 minutos"
- **5 minutos restantes**: "⏰ Quedan 5 minutos"
- **1 minuto restante**: "⏰ ¡Queda 1 minuto!"

#### Al Completar Trabajo

```
🎉 ¡Pomodoro Completado!
¡Excelente trabajo! Has completado el pomodoro #3

@usuario ¡Hora de un descanso! 🧘‍♂️
```

#### Al Iniciar Descanso

```
☕ Descanso Corto Iniciado
Hora de relajarse por 5m 0s
```

#### Al Completar Descanso

```
⏰ ¡Descanso Completado!
El tiempo de descanso ha terminado. ¿Listo para volver al trabajo?

@usuario ¡De vuelta al trabajo! 💪
```

### Comportamiento del Bot

- **Notificaciones personales**: Solo al canal donde iniciaste tu sesión
- **Menciones automáticas**: Te menciona cuando cambian las sesiones
- **Embeds coloridos**: Colores diferentes según el tipo de sesión
- **Persistencia**: Las sesiones continúan aunque te desconectes (hasta que reinicie el bot)

## 📊 Estadísticas Detalladas

### Comando `/pomodoro-stats`

```
📊 Estadísticas de Pomodoro
Estadísticas de tu sesión actual

🍅 Pomodoros: Completados: 4, Saltados: 1
☕ Descansos: Completados: 3, Saltados: 0, Descansos Largos: 1
🔥 Rachas: Actual: 2, Mejor: 4
⏱️ Tiempo Dedicado: Trabajo: 1h 40m, Descansos: 25m, Total: 2h 5m
📈 Eficiencia: 80.0% [████████████████░░░░]
📋 Info de Sesión: Total de Sesiones: 8, Iniciado: 14:30 del 1 Jun

¡Sigue con el excelente trabajo! 🎯
```

### Métricas Rastreadas

- **🍅 Pomodoros**: Completados vs saltados
- **☕ Descansos**: Cortos, largos, completados vs saltados
- **🔥 Rachas**: Pomodoros consecutivos completados
- **⏱️ Tiempo**: Total trabajado, descansado y duración de sesión
- **📈 Eficiencia**: Porcentaje de productividad con barra visual
- **📋 Historial**: Registro completo de la sesión actual

## 🏗️ Arquitectura Técnica

### Estructura del Proyecto

```
apps/discord/
├── main.go                     # Punto de entrada
├── .env.example               # Plantilla de configuración
├── .gitignore                 # Archivos ignorados
├── internal/
│   ├── bot/
│   │   ├── bot.go             # Lógica principal del bot
│   │   └── commands.go        # Manejadores de comandos slash
│   └── manager/
│       └── session_manager.go # Gestión de sesiones multi-usuario
├── go.mod                     # Dependencias
├── go.sum                     # Checksums de dependencias
└── README.md                  # Este archivo
```

### Flujo de Datos

```
Usuario Discord → Comando Slash → Bot Handler → Session Manager → Core Engine
                                      ↓
     Embed Response ← Event Handler ← Event Bus ← Core Events
```

### Componentes Clave

- **Bot**: Interfaz con Discord API, maneja comandos y respuestas
- **Session Manager**: Gestiona múltiples usuarios simultáneos
- **Event Handlers**: Convierten eventos del core a mensajes de Discord
- **Core Engine**: Motor de pomodoro thread-safe (del core)

## ⚙️ Configuración

### Variables de Entorno

| Variable                               | Descripción                        | Requerido | Defecto |
| -------------------------------------- | ---------------------------------- | --------- | ------- |
| `DISCORD_BOT_TOKEN`                    | Token del bot de Discord           | ✅        | -       |
| `DISCORD_APPLICATION_ID`               | ID de la aplicación                | ❌        | -       |
| `POMODORO_DEFAULT_WORK_DURATION`       | Duración de trabajo por defecto    | ❌        | 25m     |
| `POMODORO_DEFAULT_SHORT_BREAK`         | Descanso corto por defecto         | ❌        | 5m      |
| `POMODORO_DEFAULT_LONG_BREAK`          | Descanso largo por defecto         | ❌        | 15m     |
| `POMODORO_DEFAULT_LONG_BREAK_INTERVAL` | Pomodoros antes del descanso largo | ❌        | 4       |

### Permisos Requeridos del Bot

| Permiso              | Código     | Descripción                   |
| -------------------- | ---------- | ----------------------------- |
| Send Messages        | 2048       | Enviar mensajes y embeds      |
| Use Slash Commands   | 2147483648 | Usar comandos slash           |
| Embed Links          | 16384      | Crear embeds ricos            |
| Read Message History | 65536      | Leer contexto de mensajes     |
| Mention Everyone     | 131072     | Mencionar usuarios (opcional) |

**Código total de permisos**: `2147695616`

### Archivo .env

```env
# Configuración Requerida
DISCORD_BOT_TOKEN=tu_token_real_aquí
DISCORD_APPLICATION_ID=tu_application_id_aquí

# Configuración Opcional
POMODORO_DEFAULT_WORK_DURATION=25m
POMODORO_DEFAULT_SHORT_BREAK=5m
POMODORO_DEFAULT_LONG_BREAK=15m
POMODORO_DEFAULT_LONG_BREAK_INTERVAL=4
```

## 🚀 Despliegue

### Desarrollo Local

```bash
# Configurar entorno
cp .env.example .env
# Editar .env con tus tokens

# Instalar dependencias
go mod tidy

# Ejecutar
go run main.go
```

### Producción con Systemd

```ini
# /etc/systemd/system/gomodoro-discord.service
[Unit]
Description=Gomodoro Discord Bot
After=network.target

[Service]
Type=simple
User=tu-usuario
WorkingDirectory=/ruta/al/bot
ExecStart=/ruta/al/bot/gomodoro-discord
Restart=always
RestartSec=5
Environment=DISCORD_BOT_TOKEN=tu_token_aquí

[Install]
WantedBy=multi-user.target
```

```bash
# Habilitar y iniciar servicio
sudo systemctl enable gomodoro-discord
sudo systemctl start gomodoro-discord
sudo systemctl status gomodoro-discord
```

### Docker (Próximamente)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o discord-bot main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/discord-bot .
CMD ["./discord-bot"]
```

### Usando PM2

```bash
# Instalar PM2
npm install -g pm2

# Crear ecosystem file
cat > ecosystem.config.js << EOF
module.exports = {
  apps: [{
    name: 'gomodoro-discord',
    script: 'go',
    args: 'run main.go',
    cwd: '/ruta/al/proyecto/apps/discord',
    env: {
      DISCORD_BOT_TOKEN: 'tu_token_aquí'
    }
  }]
}
EOF

# Iniciar con PM2
pm2 start ecosystem.config.js
pm2 startup
pm2 save
```

## 🔧 Desarrollo

### Requisitos

- **Go 1.21+**
- **Token de Discord Bot**
- **Servidor de Discord para testing**

### Agregando Nuevos Comandos

1. **Definir comando** en `registerSlashCommands()` en `bot.go`
2. **Agregar case** en `handleSlashCommand()` en `bot.go`
3. **Implementar handler** en `commands.go`
4. **Actualizar documentación**

Ejemplo:

```go
// En registerSlashCommands()
{
    Name:        "pomodoro-help",
    Description: "Mostrar ayuda del bot",
},

// En handleSlashCommand()
case "pomodoro-help":
    b.handleHelpPomodoro(s, i)

// En commands.go
func (b *Bot) handleHelpPomodoro(s *discordgo.Session, i *discordgo.InteractionCreate) {
    // Implementación
}
```

### Testing

```bash
# Tests unitarios
go test ./...

# Tests con cobertura
go test -cover ./...

# Test del bot en servidor de desarrollo
# (requiere token de bot de testing)
TEST_DISCORD_TOKEN=token_test go test -v ./internal/bot/
```

### Debugging

```bash
# Logs detallados
DEBUG=true go run main.go

# Profiling de memoria
go run main.go -memprofile=mem.prof

# Análisis de performance
go tool pprof mem.prof
```

## 🐛 Troubleshooting

### Problemas Comunes

**Bot no responde a comandos:**

```bash
# Verificar que el bot esté online en Discord
# Verificar permisos del bot en el servidor
# Verificar que el token sea correcto
# Los comandos pueden tardar hasta 5 minutos en registrarse
```

**Error "Token inválido":**

```bash
# Regenerar token en Discord Developer Portal
# Verificar que no haya espacios extra en .env
# Asegurar que el token comience con "Bot " en el código
```

**Comandos no aparecen:**

```bash
# Verificar scope "applications.commands" en la invitación
# Esperar hasta 5 minutos para propagación
# Reiniciar el bot si es necesario
```

**Múltiples instancias del bot:**

```bash
# Solo una instancia del bot puede usar el mismo token
# Verificar que no haya otras instancias ejecutándose
# Usar tokens diferentes para desarrollo y producción
```

### Logs y Monitoreo

```bash
# Ver logs en tiempo real
journalctl -f -u gomodoro-discord

# Logs con PM2
pm2 logs gomodoro-discord

# Estadísticas de memoria
pm2 monit
```

## 🤝 Contribuir

### Cómo Contribuir

1. **Fork** del repositorio
2. **Crear rama** de feature (`git checkout -b feature/nueva-caracteristica`)
3. **Implementar** funcionalidad con tests
4. **Probar** con bot de desarrollo
5. **Actualizar** documentación
6. **Commit** cambios (`git commit -am 'Agrega nueva característica'`)
7. **Push** a la rama (`git push origin feature/nueva-caracteristica`)
8. **Abrir Pull Request**

### Áreas de Contribución

- 🆕 **Nuevos comandos**: Implementar funcionalidades adicionales
- 🎨 **Mejoras de UI**: Embeds más atractivos y informativos
- 📊 **Estadísticas avanzadas**: Gráficos, exportación, comparaciones
- 🔧 **Optimizaciones**: Performance, memoria, concurrencia
- 🧪 **Tests**: Aumentar cobertura de testing
- 📚 **Documentación**: Guías, ejemplos, tutorials

### Convenciones

- Usar `gofmt` para formatear código
- Seguir convenciones de Go
- Mensajes de commit descriptivos
- Tests para nueva funcionalidad
- Documentar funciones públicas

## 📄 Licencia

MIT License - ver archivo [LICENSE](../../LICENSE) para detalles.

## 🙏 Agradecimientos

- **Discord** por su excelente API y documentación
- **bwmarrin/discordgo** por la librería de Go para Discord
- **Comunidad Go** por las herramientas y soporte
- **Usuarios beta** que probaron el bot

## 📚 Enlaces Útiles

- [🏠 Proyecto Principal](../../README.md)
- [🔧 Documentación del Core](../../core/README.md)
- [🖥️ CLI App](../cli/README.md)
- [🔗 Discord Developer Portal](https://discord.com/developers/applications)
- [📖 Discord.js Guide](https://discordjs.guide/) (referencias)
- [🐛 Reportar Issues](https://github.com/tu-usuario/gomodoro/issues)

---

Hecho con ❤️ para comunidades productivas en Discord 🚀
