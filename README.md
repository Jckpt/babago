# Babago

Babago is a modern TUI (Terminal User Interface) application for downloading videos from various online platforms, using yt-dlp as the backend. The application offers an intuitive text interface with the ability to manage download presets.

## Requirements

### System Dependencies

- **Go 1.25.1** or newer
- **yt-dlp** - Main video downloading engine

### Installing yt-dlp

#### macOS (Homebrew)

```bash
brew install yt-dlp
```

#### Linux (Ubuntu/Debian)

```bash
sudo apt update
sudo apt install yt-dlp
```

## Installation

### Option 1: Download pre-built binary

1. Go to the [Releases](../../releases) section and download the appropriate file for your platform
2. Make it executable (Linux/macOS):
   ```bash
   chmod +x babago-linux-amd64
   ```
3. Move to a directory in PATH or run directly:
   ```bash
   ./babago-linux-amd64
   ```

### Option 2: Build from source

1. **Clone the repository:**

   ```bash
   git clone https://github.com/jckpt/babago.git
   cd babago
   ```

2. **Download dependencies:**

   ```bash
   go mod download
   ```

3. **Build the application:**

   ```bash
   # Build for current platform
   go build -o babago .

   # Or use Makefile
   make build
   ```

4. **Run the application:**
   ```bash
   ./babago
   ```

## Cross-platform Compilation

### Using Makefile

```bash
# Build for all platforms
make build-all

# Build for specific platforms
make build-macos    # macOS (Intel + Apple Silicon)
make build-linux    # Linux (amd64 + arm64)
make build-windows  # Windows (amd64 + 386 + arm64)

# Create ZIP archives
make package

# Clean bin directory
make clean
```

### Using bash script

```bash
# Build for all platforms
./build.sh

# Build for current platform
./build.sh current

# Build for specific platforms
./build.sh macos
./build.sh linux
./build.sh windows

# Create archives
./build.sh package

# Clean
./build.sh clean
```

### Manual cross-platform compilation

```bash
# Windows 64-bit
GOOS=windows GOARCH=amd64 go build -o babago-windows-amd64.exe .

# Linux 64-bit
GOOS=linux GOARCH=amd64 go build -o babago-linux-amd64 .

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o babago-darwin-amd64 .

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o babago-darwin-arm64 .
```

## Usage

### TUI Mode (default)

Run the application without arguments to open the TUI interface:

```bash
./babago
```

### CLI Mode

You can use babago as a wrapper for yt-dlp:

```bash
# Download from URL
./babago "https://www.youtube.com/watch?v=VIDEO_ID"

# Download with options
./babago "https://www.youtube.com/watch?v=VIDEO_ID" --format "best[height<=720]"

# Download with preset (options will be merged with saved presets)
./babago "https://www.youtube.com/watch?v=VIDEO_ID" --audio-format mp3
```

## Configuration

The application automatically creates a configuration file in the user's home directory:

- **macOS/Linux:** `~/.config/babago/config.json`

### Managing Presets

1. In TUI mode, go to the "Presets" tab (press `→`)
2. Press `N` to create a new preset
3. Enter name and yt-dlp options
4. Save preset - it will be automatically saved in configuration

## Development

### Project Structure

```
babago/
├── main.go              # Main application file
├── types.go             # Type definitions
├── config.go            # Configuration management
├── download.go          # Download logic
├── url_view.go          # URL view
├── presets_view.go      # Presets list view
├── preset_view.go       # Preset editing view
├── add_option_view.go   # Add option view
├── go.mod               # Go dependencies
├── Makefile             # Build automation
├── build.sh             # Build script
└── bin/                 # Compiled binaries
```

### Go Dependencies

- **github.com/charmbracelet/bubbletea** - TUI framework
- **github.com/charmbracelet/lipgloss** - Text styling
- **github.com/76creates/stickers** - UI components

## Troubleshooting

### Problem: "yt-dlp not found"

**Solution:** Install yt-dlp according to the instructions in the "Requirements" section.

### Problem: Application won't start

**Solution:** Check if you have the correct version of Go installed:

```bash
go version
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
