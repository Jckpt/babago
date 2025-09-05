#!/bin/bash

# Skrypt do kompilacji cross-platform dla babago

set -e

# Kolory dla output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Konfiguracja
APP_NAME="babago"
VERSION="1.0.0"
BUILD_DIR="bin"

# Funkcja do logowania
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Funkcja do tworzenia katalogu build
create_build_dir() {
    if [ ! -d "$BUILD_DIR" ]; then
        log "Tworzenie katalogu $BUILD_DIR..."
        mkdir -p "$BUILD_DIR"
    fi
}

# Funkcja do kompilacji dla konkretnej platformy
build_platform() {
    local os=$1
    local arch=$2
    local output_name="${APP_NAME}-${os}-${arch}"
    
    # Dodaj .exe dla Windows
    if [ "$os" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    log "Kompilacja dla $os/$arch..."
    
    if GOOS="$os" GOARCH="$arch" go build -ldflags="-s -w" -o "${BUILD_DIR}/${output_name}" .; then
        success "Sukces: $output_name"
    else
        error "Błąd kompilacji dla $os/$arch"
        return 1
    fi
}

# Funkcja do tworzenia archiwów
create_archives() {
    log "Tworzenie archiwów ZIP..."
    cd "$BUILD_DIR"
    
    for file in *; do
        if [ -f "$file" ]; then
            zip "${file}.zip" "$file" > /dev/null 2>&1
            success "Utworzono ${file}.zip"
        fi
    done
    
    cd ..
}

# Funkcja do wyświetlania pomocy
show_help() {
    echo "Użycie: $0 [OPCJA]"
    echo ""
    echo "Opcje:"
    echo "  all        - Kompilacja dla wszystkich platform (domyślne)"
    echo "  current    - Kompilacja dla bieżącej platformy"
    echo "  macos      - Kompilacja dla macOS (Intel + Apple Silicon)"
    echo "  linux      - Kompilacja dla Linux (amd64 + arm64)"
    echo "  windows    - Kompilacja dla Windows (amd64 + 386 + arm64)"
    echo "  package    - Tworzenie archiwów ZIP"
    echo "  clean      - Czyszczenie katalogu build"
    echo "  help       - Wyświetlenie tej pomocy"
    echo ""
    echo "Przykłady:"
    echo "  $0                    # Kompilacja dla wszystkich platform"
    echo "  $0 current           # Kompilacja dla bieżącej platformy"
    echo "  $0 macos             # Tylko macOS"
    echo "  $0 package           # Tworzenie archiwów"
}

# Funkcja do czyszczenia
clean() {
    log "Czyszczenie katalogu $BUILD_DIR..."
    rm -rf "$BUILD_DIR"
    success "Wyczyszczono!"
}

# Główna logika
main() {
    local command=${1:-all}
    
    case $command in
        "all")
            log "Kompilacja dla wszystkich platform..."
            create_build_dir
            
            # Lista platform do kompilacji
            platforms=(
                "darwin amd64"
                "darwin arm64"
                "linux amd64"
                "linux 386"
                "linux arm64"
                "linux arm"
                "windows amd64"
                "windows 386"
                "windows arm64"
                "freebsd amd64"
                "openbsd amd64"
                "netbsd amd64"
            )
            
            for platform in "${platforms[@]}"; do
                IFS=' ' read -r os arch <<< "$platform"
                build_platform "$os" "$arch"
            done
            
            success "Kompilacja dla wszystkich platform zakończona!"
            ;;
            
        "current")
            log "Kompilacja dla bieżącej platformy..."
            create_build_dir
            go build -ldflags="-s -w" -o "${BUILD_DIR}/${APP_NAME}" .
            success "Kompilacja zakończona!"
            ;;
            
        "macos")
            log "Kompilacja dla macOS..."
            create_build_dir
            build_platform "darwin" "amd64"
            build_platform "darwin" "arm64"
            success "Kompilacja macOS zakończona!"
            ;;
            
        "linux")
            log "Kompilacja dla Linux..."
            create_build_dir
            build_platform "linux" "amd64"
            build_platform "linux" "arm64"
            success "Kompilacja Linux zakończona!"
            ;;
            
        "windows")
            log "Kompilacja dla Windows..."
            create_build_dir
            build_platform "windows" "amd64"
            build_platform "windows" "386"
            build_platform "windows" "arm64"
            success "Kompilacja Windows zakończona!"
            ;;
            
        "package")
            if [ ! -d "$BUILD_DIR" ]; then
                error "Katalog $BUILD_DIR nie istnieje. Uruchom najpierw kompilację."
                exit 1
            fi
            create_archives
            ;;
            
        "clean")
            clean
            ;;
            
        "help"|"-h"|"--help")
            show_help
            ;;
            
        *)
            error "Nieznana opcja: $command"
            show_help
            exit 1
            ;;
    esac
}

# Sprawdzenie czy Go jest zainstalowane
if ! command -v go &> /dev/null; then
    error "Go nie jest zainstalowane lub nie jest w PATH"
    exit 1
fi

# Uruchomienie głównej funkcji
main "$@"
