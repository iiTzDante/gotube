package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/kkdai/youtube/v2"
)

type App struct {
	urlEntry      *widget.Entry
	statusLabel   *widget.Label
	progressBar   *widget.ProgressBar
	downloadBtn   *widget.Button
	isDownloading bool
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("GoMusic - YouTube Downloader")

	appInstance := &App{}
	appInstance.setupUI(myWindow)

	myWindow.Resize(fyne.NewSize(400, 300))
	myWindow.ShowAndRun()
}

func (a *App) setupUI(w fyne.Window) {
	// Title
	title := widget.NewLabelWithStyle(
		"🎵 GoMusic",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// URL Entry
	a.urlEntry = widget.NewEntry()
	a.urlEntry.SetPlaceHolder("Paste YouTube URL here...")
	a.urlEntry.MultiLine = false

	// Status Label
	a.statusLabel = widget.NewLabel("Ready to download")
	a.statusLabel.Wrapping = fyne.TextWrapWord

	// Progress Bar
	a.progressBar = widget.NewProgressBar()
	a.progressBar.Hide()

	// Download Button
	a.downloadBtn = widget.NewButton("Download MP3", func() {
		a.startDownload(w)
	})

	// Layout
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		widget.NewLabel("Enter YouTube URL:"),
		a.urlEntry,
		a.downloadBtn,
		widget.NewSeparator(),
		a.statusLabel,
		a.progressBar,
	)

	w.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		if k.Name == fyne.KeyReturn || k.Name == fyne.KeyEnter {
			a.startDownload(w)
		}
	})
	w.SetContent(container.NewPadded(content))
}

func (a *App) startDownload(w fyne.Window) {
	if a.isDownloading {
		return
	}

	url := strings.TrimSpace(a.urlEntry.Text)
	if url == "" {
		dialog.ShowError(fmt.Errorf("Please enter a YouTube URL"), w)
		return
	}

	videoID := extractVideoID(url)
	if videoID == "" {
		dialog.ShowError(fmt.Errorf("Invalid YouTube URL"), w)
		return
	}

	a.isDownloading = true
	a.downloadBtn.Disable()
	a.progressBar.Show()
	a.statusLabel.SetText("Fetching video info...")

	go a.downloadVideo(videoID, w)
}

func (a *App) downloadVideo(videoID string, w fyne.Window) {
	defer func() {
		fyne.Do(func() {
			a.isDownloading = false
			a.downloadBtn.Enable()
			a.progressBar.Hide()
		})
	}()

	client := youtube.Client{}

	// Get video info
	video, err := client.GetVideo(videoID)
	if err != nil {
		fyne.Do(func() {
			a.statusLabel.SetText(fmt.Sprintf("Error: %v", err))
			dialog.ShowError(err, w)
		})
		return
	}

	fyne.Do(func() {
		a.statusLabel.SetText(fmt.Sprintf("Downloading: %s", video.Title))
	})

	// Get audio format
	formats := video.Formats.Type("audio")
	if len(formats) == 0 {
		err := fmt.Errorf("No audio format found")
		fyne.Do(func() {
			a.statusLabel.SetText(err.Error())
			dialog.ShowError(err, w)
		})
		return
	}
	format := &formats[0]

	// Download audio
	tempAudio := "temp_audio"
	err = a.downloadFile(client, format, video, tempAudio)
	if err != nil {
		fyne.Do(func() {
			a.statusLabel.SetText(fmt.Sprintf("Download error: %v", err))
			dialog.ShowError(err, w)
		})
		return
	}

	// Download thumbnail
	fyne.Do(func() {
		a.statusLabel.SetText("Processing...")
	})
	tempThumb := "temp_thumb.jpg"
	_ = a.downloadThumb(video.Thumbnails[0].URL, tempThumb)

	// Convert to MP3
	finalName := sanitizeFilename(video.Title) + ".mp3"
	err = a.convertToMP3(tempAudio, tempThumb, finalName, video.Title, video.Author)
	if err != nil {
		fyne.Do(func() {
			a.statusLabel.SetText(fmt.Sprintf("Conversion error: %v", err))
			dialog.ShowError(err, w)
		})
		return
	}

	// Cleanup
	os.Remove(tempAudio)
	os.Remove(tempThumb)

	fyne.Do(func() {
		a.statusLabel.SetText(fmt.Sprintf("✓ Saved: %s", finalName))
		a.progressBar.SetValue(0)
		dialog.ShowInformation("Success", fmt.Sprintf("Downloaded:\n%s", finalName), w)
	})
}

func (a *App) downloadFile(client youtube.Client, format *youtube.Format, video *youtube.Video, path string) error {
	stream, size, err := client.GetStream(video, format)
	if err != nil {
		return err
	}
	defer stream.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			file.Write(buf[:n])
			downloaded += int64(n)
			if size > 0 {
				progress := float64(downloaded) / float64(size)
				fyne.Do(func() {
					a.progressBar.SetValue(progress)
				})
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) downloadThumb(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func (a *App) convertToMP3(audioPath, thumbPath, outputPath, title, artist string) error {
	args := []string{
		"-y",
		"-i", audioPath,
		"-i", thumbPath,
		"-map", "0:0",
		"-map", "1:0",
		"-c:a", "libmp3lame",
		"-q:a", "2",
		"-id3v2_version", "3",
		"-metadata:s:v", `title="Album cover"`,
		"-metadata:s:v", `comment="Cover (Front)"`,
		"-metadata", "title=" + title,
		"-metadata", "artist=" + artist,
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	return cmd.Run()
}

func extractVideoID(url string) string {
	patterns := []string{
		`(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/shorts/)([a-zA-Z0-9_-]{11})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	return name
}
