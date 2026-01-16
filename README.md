# GoTube

A beautiful terminal UI for downloading YouTube videos with format selection and HD support.

## Features

- 🎬 Interactive TUI built with Bubble Tea
- 📺 HD video support (up to 4K)
- 🎵 Multi-audio track selection with language info
- 📊 Real-time download progress
- 🔄 Automatic video+audio merging with FFmpeg
- 🎨 Clean, modern interface

## Installation

```bash
go install github.com/iiTzDante/gotube@latest
```

Or build from source:

```bash
git clone https://github.com/iiTzDante/gotube
cd gotube
go build -o gotube main.go
```

## Requirements

- Go 1.19 or later
- FFmpeg (for HD video merging)

## Usage

```bash
gotube <youtube-url>
```

Example:
```bash
gotube "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
```

## How It Works

1. Fetches video information and available formats
2. Displays interactive format selection menu
3. For HD videos: downloads video and audio separately
4. Merges streams using FFmpeg
5. Saves final MP4 file

## Dependencies

- [kkdai/youtube](https://github.com/kkdai/youtube) - YouTube video fetching
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## License

MIT
