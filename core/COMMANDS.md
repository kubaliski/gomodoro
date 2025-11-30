# 🧪 Comandos de Testing - Gomodoro Core

Comandos disponibles para desarrollo y testing.

## 🚀 Setup Inicial

```bash
# Hacer script ejecutable
chmod +x scripts/test.sh

# Verificar setup
./scripts/test.sh setup
```

## 🧪 Ejecutar Tests

```bash
# Tests básicos
./scripts/test.sh test

# Tests con race detection
./scripts/test.sh test-race

# Tests con cobertura
./scripts/test.sh test-cover

# Limpiar archivos generados
./scripts/test.sh clean
```

## 📦 Tests por Paquete

```bash
# Tests específicos
go test -v ./config
go test -v ./timer
go test -v ./engine
go test -v ./events
go test -v ./stats
go test -v ./testhelpers
```

## 📊 Comandos de Cobertura

```bash
# Generar reporte de cobertura
go test -coverprofile=coverage.out ./...

# Ver cobertura total
go tool cover -func=coverage.out | grep total

# Generar reporte HTML
go tool cover -html=coverage.out
```

## 🏁 Desarrollo Rápido

```bash
# Workflow completo
./scripts/test.sh setup      # Una vez
./scripts/test.sh test       # Durante desarrollo
./scripts/test.sh test-race  # Antes de commit
```

---

**Estado**: ✅ Tarea 1 completada - Setup de herramientas de testing
**Siguiente**: 🎯 Tarea 2 - Config Package Tests
