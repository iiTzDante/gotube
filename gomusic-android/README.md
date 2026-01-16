# GoMusic Android - Fyne GUI Version

A cross-platform GUI application for downloading YouTube music, built with Fyne framework. Can be packaged as an Android APK.

## Features

- ✅ **URL-based downloads** - Paste any YouTube URL
- ✅ **High-quality audio** - Downloads best available audio format
- ✅ **MP3 conversion** - Automatic conversion with FFmpeg
- ✅ **ID3 tags** - Embeds title, artist, and cover art
- ✅ **Progress tracking** - Real-time download progress
- ✅ **Cross-platform** - Works on Linux, Windows, macOS, and Android

## Requirements

- Go 1.19 or later
- FFmpeg (must be in PATH)
- For Android: Android SDK and NDK

## Quick Start

### Desktop Usage

```bash
cd /home/cesario/Documents/gotube/gomusic-android
./gomusic-gui
```

### Build from Source

```bash
go build -o gomusic-gui main.go
```

## Building for Android

### Prerequisites

1. Install Android Studio
2. Install Android SDK and NDK via SDK Manager
3. Set environment variable:
   ```bash
   export ANDROID_NDK_HOME=/path/to/android/ndk
   ```

### Install Fyne CLI

```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
```

### Build APK

```bash
fyne package -os android -appID com.gomusic.app -icon icon.png
```

Or use `fyne-cross` for easier cross-compilation:

```bash
go install github.com/fyne-io/fyne-cross@latest
fyne-cross android -app-id com.gomusic.app
```

The APK will be generated in the current directory.

## Usage

1. Launch the app
2. Paste a YouTube URL (supports youtube.com, youtu.be, shorts)
3. Click "Download MP3"
4. Wait for download and conversion
5. Find your MP3 in the current directory

## Supported URL Formats

- `https://www.youtube.com/watch?v=VIDEO_ID`
- `https://youtu.be/VIDEO_ID`
- `https://www.youtube.com/shorts/VIDEO_ID`

## Project Structure

```
gomusic-android/
├── main.go          # Main application code
├── FyneApp.toml     # Fyne metadata
├── go.mod           # Go dependencies
├── go.sum           # Dependency checksums
└── README.md        # This file
```

## Differences from TUI Version

| Feature | TUI (gomusic) | GUI (gomusic-android) |
|---------|---------------|----------------------|
| Search | ✅ Yes | ❌ No (URL only) |
| Platform | Terminal | Desktop + Android |
| UI | Text-based | Graphical |
| Dependencies | Browser scraper | None (uses kkdai/youtube) |

## Future Enhancements

- [ ] Add search functionality via backend API
- [ ] Batch download support
- [ ] Download history
- [ ] Custom output directory selection
- [ ] Quality selection (bitrate)

## Troubleshooting

### "FFmpeg not found"
Ensure FFmpeg is installed and in your PATH:
```bash
ffmpeg -version
```

### Android build fails
- Verify `ANDROID_NDK_HOME` is set correctly
- Ensure Android SDK platform-tools are installed
- Try using `fyne-cross` instead of `fyne package`

## License

Same as parent project (GoTube/GoMusic)
