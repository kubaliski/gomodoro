# Pomodoro CLI

Un temporizador Pomodoro simple para la línea de comandos, escrito en Go. Proyecto para trabajar con las Goroutines y aprender

## Instalación

```bash
git clone <tu-repo>
cd pomodoro-cli
go build -o pomodoro
```

## Uso

```bash
./pomodoro
```

## Características (Planificadas)

- ⏱️ Timer de 25 minutos por defecto
- ⏸️ Pausa y reanuda
- 🔄 Ciclos automáticos trabajo/descanso
- 📊 Estadísticas básicas
- 🎨 Interfaz colorida en terminal

## Desarrollo

Este proyecto está siendo desarrollado como práctica de Go, siguiendo un enfoque incremental.

### Estructura del proyecto

```
pomodoro-cli/
├── main.go                 # Punto de entrada
├── internal/
│   ├── timer/             # Lógica del temporizador
│   ├── config/            # Configuración
│   └── ui/                # Interfaz de usuario
└── README.md
```

## Roadmap

- [ ] Timer básico funcional
- [ ] Interfaz de usuario mejorada
- [ ] Control interactivo (pause/resume)
- [ ] Ciclo completo de pomodoros
- [ ] Persistencia de estadísticas
