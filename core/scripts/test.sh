#!/bin/bash
# test.sh - Script para verificar y ejecutar tests (consistente con testing-scripts.sh del bot)

set -e

echo "🧪 Pomodoro Core - Testing Scripts"
echo "=================================="

# Función para mostrar uso
show_usage() {
    echo "Usage: $0 {setup|test|test-race|test-cover|clean}"
    echo ""
    echo "Commands:"
    echo "  setup      - Verificar setup de testing"
    echo "  test       - Ejecutar todos los tests"
    echo "  test-race  - Ejecutar tests con race detection"
    echo "  test-cover - Ejecutar tests con cobertura"
    echo "  clean      - Limpiar archivos generados"
    echo ""
    echo "Quick start:"
    echo "1. ./scripts/test.sh setup"
    echo "2. ./scripts/test.sh test"
    echo "3. ./scripts/test.sh test-race"
}

# Función para verificar setup
check_setup() {
    echo "🔧 Verificando setup de testing..."

    # Verificar go.mod
    if [ ! -f "go.mod" ]; then
        echo "❌ go.mod no encontrado"
        exit 1
    fi
    echo "✅ go.mod encontrado"

    # Descargar dependencias
    echo "📦 Descargando dependencias..."
    go mod download
    go mod tidy
    echo "✅ Dependencias actualizadas"

    # Verificar estructura testhelpers
    if [ ! -d "testhelpers" ]; then
        echo "❌ Directorio testhelpers no encontrado"
        exit 1
    fi

    required_files=(
        "testhelpers/helpers.go"
        "testhelpers/mocks.go"
        "testhelpers/fixtures.go"
        "testhelpers/testhelpers_test.go"
    )

    echo "📁 Verificando archivos requeridos..."
    for file in "${required_files[@]}"; do
        if [ -f "$file" ]; then
            echo "✅ $file"
        else
            echo "❌ $file no encontrado"
            exit 1
        fi
    done

    # Verificar que compila
    echo "🔨 Verificando compilación..."
    go build ./...
    echo "✅ Código compila correctamente"

    # Ejecutar test de verificación básico
    echo "🧪 Ejecutando tests de verificación..."
    go test -v ./testhelpers -run TestSetupVerification
    echo "✅ Tests de verificación completados"

    echo ""
    echo "🎉 Setup verificado correctamente"
    echo "📋 Próximos pasos:"
    echo "   ./scripts/test.sh test       - Ejecutar todos los tests"
    echo "   ./scripts/test.sh test-race  - Con race detection"
    echo "   ./scripts/test.sh test-cover - Con cobertura"
}

# Función para ejecutar tests
run_tests() {
    echo "🧪 Ejecutando todos los tests..."
    go test ./...
    echo "✅ Tests completados"
}

# Función para tests con race detection
run_tests_race() {
    echo "🧪 Ejecutando tests con race detection..."

    # Verificar si gcc está disponible
    if ! command -v gcc >/dev/null 2>&1; then
        echo "⚠️  GCC no encontrado. Race detection requiere un compilador C."
        echo "📋 Opciones:"
        echo "   1. Instalar TDM-GCC: https://jmeubank.github.io/tdm-gcc/"
        echo "   2. Instalar con chocolatey: choco install mingw"
        echo "   3. Usar WSL para race detection"
        echo ""
        echo "🔄 Ejecutando tests normales en su lugar..."
        go test ./...
        echo "⚠️  Tests ejecutados SIN race detection"
        return
    fi

    # Si CGO está deshabilitado, habilitarlo temporalmente
    if [ "${CGO_ENABLED:-1}" = "0" ]; then
        echo "🔧 Habilitando CGO temporalmente para race detection..."
        CGO_ENABLED=1 go test -race ./...
    else
        go test -race ./...
    fi

    echo "✅ Tests con race detection completados"
}

# Función para tests con cobertura
run_tests_cover() {
    echo "🧪 Ejecutando tests con cobertura..."
    go test -coverprofile=coverage.out ./...
    echo ""
    echo "📊 Resumen de cobertura:"
    go tool cover -func=coverage.out | grep total
    echo ""
    echo "💡 Para ver reporte HTML: go tool cover -html=coverage.out"
    echo "✅ Tests con cobertura completados"
}

# Función para limpiar archivos
clean_files() {
    echo "🧹 Limpiando archivos generados..."
    rm -f coverage.out
    echo "✅ Archivos limpiados"
}

# Función principal
case "$1" in
    "setup")
        check_setup
        ;;
    "test")
        run_tests
        ;;
    "test-race")
        run_tests_race
        ;;
    "test-cover")
        run_tests_cover
        ;;
    "clean")
        clean_files
        ;;
    *)
        show_usage
        ;;
esac