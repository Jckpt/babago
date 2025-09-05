# Makefile dla babago - kompilacja cross-platform

# Nazwa aplikacji
APP_NAME = babago

# Wersja (możesz zmienić)
VERSION = 1.0.0

# Katalog na binarne pliki
BUILD_DIR = bin

# Definicje platform i architektur
PLATFORMS = \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/386 \
	linux/arm64 \
	linux/arm \
	windows/amd64 \
	windows/386 \
	windows/arm64 \
	freebsd/amd64 \
	openbsd/amd64 \
	netbsd/amd64

# Domyślny target
.PHONY: all
all: clean build-all

# Tworzenie katalogu bin
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Kompilacja dla wszystkich platform
.PHONY: build-all
build-all: $(BUILD_DIR)
	@echo "Kompilacja dla wszystkich platform..."
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		OUTPUT_NAME=$(APP_NAME)-$$OS-$$ARCH; \
		if [ $$OS = "windows" ]; then \
			OUTPUT_NAME=$$OUTPUT_NAME.exe; \
		fi; \
		echo "Kompilacja dla $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH go build -ldflags="-s -w" -o $(BUILD_DIR)/$$OUTPUT_NAME .; \
	done
	@echo "Kompilacja zakończona!"

# Kompilacja dla bieżącej platformy
.PHONY: build
build: $(BUILD_DIR)
	@echo "Kompilacja dla bieżącej platformy..."
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) .
	@echo "Kompilacja zakończona!"

# Kompilacja dla macOS (Intel + Apple Silicon)
.PHONY: build-macos
build-macos: $(BUILD_DIR)
	@echo "Kompilacja dla macOS..."
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 .
	@echo "Kompilacja macOS zakończona!"

# Kompilacja dla Linux
.PHONY: build-linux
build-linux: $(BUILD_DIR)
	@echo "Kompilacja dla Linux..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 .
	@echo "Kompilacja Linux zakończona!"

# Kompilacja dla Windows
.PHONY: build-windows
build-windows: $(BUILD_DIR)
	@echo "Kompilacja dla Windows..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=386 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-windows-386.exe .
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-windows-arm64.exe .
	@echo "Kompilacja Windows zakończona!"

# Tworzenie archiwów ZIP
.PHONY: package
package: build-all
	@echo "Tworzenie archiwów..."
	@cd $(BUILD_DIR) && for file in *; do \
		if [ -f "$$file" ]; then \
			zip "$$file.zip" "$$file"; \
			echo "Utworzono $$file.zip"; \
		fi; \
	done
	@echo "Archiwa utworzone!"

# Czyszczenie
.PHONY: clean
clean:
	@echo "Czyszczenie katalogu bin..."
	rm -rf $(BUILD_DIR)
	@echo "Wyczyszczono!"

# Instalacja zależności
.PHONY: deps
deps:
	@echo "Pobieranie zależności..."
	go mod download
	go mod tidy
	@echo "Zależności pobrane!"

# Testy
.PHONY: test
test:
	@echo "Uruchamianie testów..."
	go test ./...
	@echo "Testy zakończone!"

# Sprawdzenie kodu
.PHONY: lint
lint:
	@echo "Sprawdzanie kodu..."
	go vet ./...
	gofmt -l .
	@echo "Sprawdzanie zakończone!"

# Pomoc
.PHONY: help
help:
	@echo "Dostępne komendy:"
	@echo "  make build          - Kompilacja dla bieżącej platformy"
	@echo "  make build-all      - Kompilacja dla wszystkich platform"
	@echo "  make build-macos    - Kompilacja dla macOS (Intel + Apple Silicon)"
	@echo "  make build-linux    - Kompilacja dla Linux (amd64 + arm64)"
	@echo "  make build-windows  - Kompilacja dla Windows (amd64 + 386 + arm64)"
	@echo "  make package        - Tworzenie archiwów ZIP"
	@echo "  make clean          - Czyszczenie katalogu build"
	@echo "  make deps           - Pobieranie zależności"
	@echo "  make test           - Uruchamianie testów"
	@echo "  make lint           - Sprawdzanie kodu"
	@echo "  make help           - Wyświetlenie tej pomocy"
