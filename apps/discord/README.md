# 💬 Gomodoro Discord Bot

Bot de Discord que lleva la Técnica Pomodoro a tu servidor usando el motor central de Gomodoro. Diseñado para equipos y comunidades que quieren mejorar su productividad de manera colaborativa.

## ✨ Características

- **📱 Notificaciones privadas inteligentes**: Mensajes DM automáticos con fallback transparente al canal
- **👥 Sesiones simultáneas**: Cada usuario puede tener su propia sesión independiente
- **🎛️ Timers personalizables**: Configura duración de trabajo y descansos por sesión
- **🎨 Integración rica con Discord**: Embeds hermosos, comandos slash y notificaciones
- **📊 Seguimiento de estadísticas**: Rastrea tu productividad a través de sesiones
- **⏸️ Control completo**: Pausa, reanuda y salta sesiones
- **🔔 Sistema robusto de notificaciones**: Sin configuración, funciona automáticamente
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

## 📱 Sistema de Notificaciones Inteligentes

### 🎯 Funcionamiento Automático

El bot ahora utiliza un **sistema de notificaciones inteligente** que funciona automáticamente sin configuración:

#### ✅ **Para Nuevos Usuarios:**

1. Ejecutas `/pomodoro` → **Respuesta pública** en el canal
2. **Mensaje de bienvenida automático** en tus mensajes privados
3. **Todas las notificaciones posteriores** van a DM (mensajes privados)

#### 🔄 **Fallback Inteligente:**

- Si tienes **DMs deshabilitados** → Las notificaciones van al canal automáticamente
- Si **bloqueas el bot** → Fallback transparente al canal público
- **Sin errores visibles** → Todo funciona sin problemas

#### 📍 **Dónde van las notificaciones:**

| Tipo de Mensaje                | Ubicación                | Descripción                        |
| ------------------------------ | ------------------------ | ---------------------------------- |
| **Comandos** (respuestas)      | 📢 **Canal público**     | `/pomodoro`, `/status`, etc.       |
| **Notificaciones** de pomodoro | 📱 **Mensajes privados** | Completado, inicios, recordatorios |
| **Fallback** (si DM falla)     | 📢 **Canal público**     | Automático y transparente          |

### 🔔 Tipos de Notificaciones

#### Durante el Trabajo (25 min por defecto)

```
📱 Mensaje Privado:
⏰ Quedan 5 minutos para completar el pomodoro
¡Sigue enfocado! 💪
```

#### Al Completar Trabajo

```
📱 Mensaje Privado:
🎉 ¡Pomodoro Completado!
¡Excelente trabajo! Has completado el pomodoro #3
¡Hora de un descanso! 🧘‍♂️
```

#### Al Iniciar Descanso

```
📱 Mensaje Privado:
☕ Descanso Corto Iniciado
Hora de relajarse por 5m 0s
```

#### Mensaje de Bienvenida (Primera vez)

```
📱 Mensaje Privado:
👋 ¡Bienvenido al sistema de notificaciones de Gomodoro!

🔔 Recibirás notificaciones privadas sobre:
• Inicio y finalización de pomodoros
• Recordatorios de tiempo restante
• Cambios de estado de tu sesión

Los comandos siempre responden en el canal público donde los ejecutes.

¡Que tengas una sesión productiva! 🍅
```

### ⚙️ Ventajas del Sistema

#### ✅ **Sin Configuración**

- **Funciona automáticamente** para todos los usuarios
- **No necesitas comandos especiales** de configuración
- **Experiencia consistente** en todos los servidores

#### ✅ **Privacidad Mejorada**

- **Notificaciones personales** no interrumpen canales públicos
- **Canales limpios** sin spam de notificaciones
- **Cada usuario** recibe solo sus notificaciones

#### ✅ **Robustez Total**

- **Fallback automático** cuando DMs no están disponibles
- **Sin errores visibles** para el usuario
- **Funciona en cualquier configuración** de Discord

#### ✅ **Multi-Usuario Perfecto**

- **Aislamiento completo** entre usuarios
- **Cero crosstalk** entre sesiones
- **Performance optimizada** con cache inteligente

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
├── testing-scripts.sh         # Scripts de testing y monitoreo
├── internal/
│   └── bot/
│       ├── bot.go             # Core del bot y configuración
│       ├── commands.go        # Manejadores de comandos slash
│       ├── notifications.go   # Sistema DM con fallback inteligente
│       ├── events.go          # Event handlers para pomodoro
│       ├── registry.go        # Registro de comandos slash
│       ├── utils.go           # Funciones helper compartidas
│       └── session_manager.go # Gestión de sesiones multi-usuario
├── go.mod                     # Dependencias
├── go.sum                     # Checksums de dependencias
└── README.md                  # Este archivo
```

### Arquitectura Modular

```
Bot Core                    Sistema de Notificaciones
├── Commands Handler        ├── DM Manager (cache inteligente)
├── Event System           ├── Fallback System
├── Session Manager        ├── Welcome Messages
└── Slash Registry         └── Multi-User Isolation

       ↓                           ↓
   Session Engine    →    Event Bus    →    Discord API
```

### Flujo de Notificaciones

```
Pomodoro Event → Event Handler → Notification Manager → DM Channel (cached)
                                        ↓ (si falla)
                                 Channel Fallback → Public Channel
```

### Componentes Clave

- **Bot Core**: Interfaz con Discord API, lifecycle management
- **Notification Manager**: Sistema DM inteligente con fallback automático
- **Event Handlers**: Convierten eventos del core a mensajes formateados
- **Session Manager**: Gestiona múltiples usuarios con cache DM
- **Command Registry**: Registro centralizado de comandos slash
- **Utils**: Funciones helper compartidas entre componentes

## 🧪 Testing y Monitoreo

### Script de Testing Automatizado

El proyecto incluye un script de testing completo que funciona en Windows, macOS y Linux:

```bash
# Verificar configuración
./testing-scripts.sh setup

# Iniciar bot con logging automático
./testing-scripts.sh start

# Monitorear logs en tiempo real (otra terminal)
./testing-scripts.sh monitor

# Analizar resultados de testing
./testing-scripts.sh analyze

# Parar el bot
./testing-scripts.sh stop
```

### Funciones del Script

| Comando    | Descripción                            |
| ---------- | -------------------------------------- |
| `setup`    | Verificar configuración y compilar bot |
| `start`    | Iniciar bot con logging automático     |
| `stop`     | Parar bot de forma limpia              |
| `restart`  | Reiniciar bot (útil para desarrollo)   |
| `monitor`  | Logs en tiempo real con colores        |
| `analyze`  | Análisis de métricas y errores         |
| `commands` | Guía de comandos para testing          |
| `clean`    | Limpiar logs para testing fresco       |

### Logs Inteligentes

El sistema genera logs descriptivos con emojis para facilitar el debugging:

```bash
✅ Bot is ready! Logged in as: Gomodoro#7460
🚀 Starting new session for user...
📱 Created and cached DM channel for user...
👋 Welcome message sent to user...
📱 DM notification sent successfully to user...
📢 DM unavailable, using channel fallback... (si es necesario)
```

### Métricas de Performance

El script de análisis proporciona métricas en tiempo real:

```
=== TEST RESULTS SUMMARY ===
📊 Sessions Started: 5
📱 DM Notifications Sent: 23
📢 Fallback Notifications: 0
👋 Welcome Messages: 2
❌ Errors: 0
📈 DM Success Rate: 100%
```

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

| Permiso              | Código     | Descripción               |
| -------------------- | ---------- | ------------------------- |
| Send Messages        | 2048       | Enviar mensajes y embeds  |
| Use Slash Commands   | 2147483648 | Usar comandos slash       |
| Embed Links          | 16384      | Crear embeds ricos        |
| Read Message History | 65536      | Leer contexto de mensajes |
| Send Messages in DM  | -          | Enviar mensajes privados  |

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

# Opción 1: Ejecutar directamente
go run main.go

# Opción 2: Usar scripts de testing
./testing-scripts.sh start
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

1. **Definir comando** en `registerSlashCommands()` en `registry.go`
2. **Agregar case** en `handleSlashCommand()` en `bot.go`
3. **Implementar handler** en `commands.go`
4. **Actualizar documentación**

Ejemplo:

```go
// En registry.go - registerSlashCommands()
{
    Name:        "pomodoro-help",
    Description: "Mostrar ayuda del bot",
},

// En bot.go - handleSlashCommand()
case "pomodoro-help":
    b.handleHelpPomodoro(s, i)

// En commands.go
func (b *Bot) handleHelpPomodoro(s *discordgo.Session, i *discordgo.InteractionCreate) {
    // Implementación
}
```

### Testing y Debugging

```bash
# Testing automatizado con script
./testing-scripts.sh setup    # Verificar configuración
./testing-scripts.sh start    # Iniciar con logging
./testing-scripts.sh monitor  # Monitorear en tiempo real
./testing-scripts.sh analyze  # Analizar resultados

# Testing manual
go test ./...

# Testing con cobertura
go test -cover ./...

# Debugging con logs detallados
DEBUG=true go run main.go
```

### Workflow de Desarrollo

```bash
# 1. Hacer cambios en el código
vim internal/bot/commands.go

# 2. Reiniciar bot para probar
./testing-scripts.sh restart

# 3. Monitorear logs para debugging
./testing-scripts.sh monitor

# 4. Analizar métricas
./testing-scripts.sh analyze

# 5. Commit cuando esté listo
git add .
git commit -m "feat: nueva funcionalidad"
```

## 🐛 Troubleshooting

### Problemas Comunes

#### **Bot no responde a comandos:**

```bash
# Verificar que el bot esté online en Discord
./testing-scripts.sh analyze  # Ver si hay errores

# Verificar permisos del bot en el servidor
# Los comandos pueden tardar hasta 5 minutos en registrarse
```

#### **No recibo notificaciones en DM:**

**Esto es completamente normal si:**

- Tienes DMs deshabilitados para el servidor
- Has bloqueado el bot
- Tienes configuraciones de privacidad restrictivas

**✅ Solución automática:**

- El bot detecta esto y envía notificaciones al canal público
- No necesitas hacer nada, funciona automáticamente
- En los logs verás: `📢 DM unavailable, using channel fallback`

#### **Las notificaciones van al canal público:**

**Esto significa que:**

- ✅ El bot está funcionando correctamente
- ✅ El fallback automático está activado
- ✅ Tus DMs pueden estar deshabilitados

**Para recibir DMs:**

1. Habilita "Permitir mensajes privados de miembros del servidor"
2. Reinicia tu sesión con `/pomodoro-stop` y `/pomodoro`

#### **Error "Token inválido":**

```bash
# Verificar .env
cat .env  # Debe tener DISCORD_BOT_TOKEN=...

# Regenerar token en Discord Developer Portal
# Verificar que no haya espacios extra en .env
```

#### **Comandos no aparecen:**

```bash
# Verificar scope "applications.commands" en la invitación
# Esperar hasta 5 minutos para propagación
./testing-scripts.sh restart  # Reiniciar el bot
```

### Análisis de Logs

```bash
# Ver logs en tiempo real con colores
./testing-scripts.sh monitor

# Buscar errores específicos
grep "❌" logs/bot.log

# Buscar fallbacks de DM
grep "📢.*fallback" logs/bot.log

# Ver resumen completo
./testing-scripts.sh analyze
```

### Estados del Sistema

#### ✅ **Sistema Funcionando Correctamente:**

```
📱 DM Success Rate: 100%
❌ Errors: 0
📈 Performance: <1s response time
```

#### 🔄 **Sistema con Fallback (Normal):**

```
📱 DM Success Rate: 70%
📢 Fallback Notifications: 30%
❌ Errors: 0
```

#### ❌ **Sistema con Problemas:**

```
❌ Errors: >0
🚫 Bot offline or token issues
⏱️ Performance: >5s response time
```

## 🤝 Contribuir

### Cómo Contribuir

1. **Fork** del repositorio
2. **Crear rama** de feature (`git checkout -b feature/nueva-caracteristica`)
3. **Implementar** funcionalidad con tests
4. **Probar** con script de testing (`./testing-scripts.sh`)
5. **Actualizar** documentación
6. **Commit** cambios (`git commit -am 'feat: nueva característica'`)
7. **Push** a la rama (`git push origin feature/nueva-caracteristica`)
8. **Abrir Pull Request**

### Áreas de Contribución

- 🆕 **Nuevos comandos**: Implementar funcionalidades adicionales
- 🎨 **Mejoras de UI**: Embeds más atractivos y informativos
- 📊 **Estadísticas avanzadas**: Gráficos, exportación, comparaciones
- 🔧 **Optimizaciones**: Performance, memoria, concurrencia
- 🧪 **Tests**: Aumentar cobertura de testing automatizado
- 📚 **Documentación**: Guías, ejemplos, tutorials

### Convenciones

- Usar `gofmt` para formatear código
- Seguir convenciones de Go
- Mensajes de commit descriptivos
- Probar con `./testing-scripts.sh` antes de commit
- Tests para nueva funcionalidad
- Documentar funciones públicas

## 📄 Licencia

MIT License - ver archivo [LICENSE](../../LICENSE) para detalles.

## 🙏 Agradecimientos

- **Discord** por su excelente API y documentación
- **bwmarrin/discordgo** por la librería de Go para Discord
- **Comunidad Go** por las herramientas y soporte
- **Usuarios beta** que probaron el sistema de notificaciones DM

## 📚 Enlaces Útiles

- [🏠 Proyecto Principal](../../README.md)
- [🔧 Documentación del Core](../../core/README.md)
- [🖥️ CLI App](../cli/README.md)
- [🔗 Discord Developer Portal](https://discord.com/developers/applications)
- [📖 Discord.js Guide](https://discordjs.guide/) (referencias)
- [🐛 Reportar Issues](https://github.com/tu-usuario/gomodoro/issues)

---

**Versión 0.1.0** - Ahora con notificaciones privadas inteligentes y arquitectura modular
