# GoTube

A beautiful, high-performance terminal UI for searching and downloading YouTube videos with full HD support.

## Features

- 🎬 **Search & Download**: Integrated search for videos without high-overhead browsers.
- 📺 **Ultra HD Support**: Downloads and merges video/audio for 1080p, 1440p, and 4K support.
- 🎵 **Multi-Track Audio**: Detailed selection for multi-language or high-quality audio tracks.
- 📊 **Real-Time Progress**: Smooth, accurate progress bars for all downloads.
- 🔄 **Auto-Merge**: Intelligent stream merging using FFmpeg.
- 🎨 **Sleek TUI**: Modern interface built for speed and clarity.

## Installation

### Arch Linux (AUR)
```bash
yay -S gotube
```

### From Source
```bash
git clone https://github.com/iiTzDante/gotube
cd gotube
go build -o gotube .
```

## Requirements

- **Go 1.22+**
- **FFmpeg** (Required for merging high-definition streams)

## Usage

```bash
gotube <youtube-url-or-search-query>
```

## Controls

| Key | Action |
|-----|--------|
| `Enter` | Confirm Selection |
| `↑` / `↓` | Navigate Menus |
| `q` | Back / Quit |
| `Ctrl+C` | Force Quit |

## How It Works

1.  **Metadata Fetch**: Retrieves stream URLs and format details efficiently.
2.  **Smart Selection**: Lets you choose the exact resolution and audio quality.
3.  **Parallel Streams**: Downloads video and audio chunks concurrently.
4.  **FFmpeg Fusion**: Seamlessly merges streams into a final MP4 container.

## Dependencies

- [kkdai/youtube](https://github.com/kkdai/youtube) - Stream fetching engine
- [raitonoberu/ytsearch](https://github.com/raitonoberu/ytsearch) - High-speed search
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework

## License

MIT
