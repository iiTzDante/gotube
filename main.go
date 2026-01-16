package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kkdai/youtube/v2"
)

// --- Styles ---

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginTop(1).
			MarginBottom(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

// --- Types ---

type state int

const (
	stateFetching state = iota
	stateSelectingVideo
	stateSelectingAudio
	stateDownloading
	stateMerging
	stateFinished
	stateError
)

type formatItem struct {
	format *youtube.Format
	title  string
	desc   string
}

func (i formatItem) Title() string       { return i.title }
func (i formatItem) Description() string { return i.desc }
func (i formatItem) FilterValue() string { return i.title }

type model struct {
	state         state
	videoURL      string
	video         *youtube.Video
	list          list.Model
	progress      progress.Model
	spinner       spinner.Model
	err           error
	fileName      string
	quitting      bool
	program       *tea.Program
	width         int
	height        int
	progressV     float64
	progressA     float64
	selectedVideo *youtube.Format
	selectedAudio *youtube.Format
}

// --- Messages ---

type videoInfoFetchedMsg *youtube.Video
type errMsg error
type downloadProgressMsg float64
type mergingMsg struct{}
type downloadDoneMsg string

// --- Logic ---

func fetchVideoInfo(url string) tea.Cmd {
	return func() tea.Msg {
		client := youtube.Client{}
		video, err := client.GetVideo(url)
		if err != nil {
			return errMsg(err)
		}
		return videoInfoFetchedMsg(video)
	}
}

type progressReader struct {
	io.Reader
	total      int64
	downloaded int64
	onProgress func(float64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.downloaded += int64(n)
	if pr.total > 0 {
		pr.onProgress(float64(pr.downloaded) / float64(pr.total))
	}
	return n, err
}

func (m *model) runDownload() {
	client := youtube.Client{}

	filename := strings.ReplaceAll(m.video.Title, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")

	if m.selectedAudio == nil {
		// Normal download (combined)
		err := m.downloadSingle(client, m.selectedVideo, filename+".mp4", func(p float64) {
			m.program.Send(downloadProgressMsg(p))
		})
		if err != nil {
			m.program.Send(errMsg(err))
			return
		}
		m.fileName = filename + ".mp4"
	} else {
		// DASH download: Video + Audio separately
		videoFile := filename + ".temp.mp4"
		audioFile := filename + ".temp.m4a"
		finalFile := filename + ".mp4"

		// Download Video
		err := m.downloadSingle(client, m.selectedVideo, videoFile, func(p float64) {
			m.progressV = p
			m.program.Send(downloadProgressMsg((m.progressV + m.progressA) / 2))
		})
		if err != nil {
			m.program.Send(errMsg(err))
			return
		}

		// Download Audio
		err = m.downloadSingle(client, m.selectedAudio, audioFile, func(p float64) {
			m.progressA = p
			m.program.Send(downloadProgressMsg((m.progressV + m.progressA) / 2))
		})
		if err != nil {
			m.program.Send(errMsg(err))
			return
		}

		// Merge
		m.program.Send(mergingMsg{})

		cmd := exec.Command("ffmpeg", "-y", "-i", videoFile, "-i", audioFile, "-c", "copy", finalFile)
		if err := cmd.Run(); err != nil {
			m.program.Send(errMsg(fmt.Errorf("merging failed: %v (is ffmpeg installed?)", err)))
			return
		}

		os.Remove(videoFile)
		os.Remove(audioFile)
		m.fileName = finalFile
	}

	m.program.Send(downloadDoneMsg(m.fileName))
}

func (m *model) downloadSingle(client youtube.Client, format *youtube.Format, path string, onProgress func(float64)) error {
	stream, size, err := client.GetStream(m.video, format)
	if err != nil {
		return err
	}
	defer stream.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	pr := &progressReader{
		Reader:     stream,
		total:      size,
		onProgress: onProgress,
	}

	_, err = io.Copy(file, pr)
	return err
}

// --- Bubble Tea Methods ---

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchVideoInfo(m.videoURL),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.state == stateSelectingVideo {
				item, ok := m.list.SelectedItem().(formatItem)
				if ok {
					m.selectedVideo = item.format
					if item.format.AudioChannels == 0 && strings.Contains(item.format.MimeType, "video") {
						// HD format, need to select audio
						m.state = stateSelectingAudio
						audioFormats := m.video.Formats.Type("audio")
						var items []list.Item
						for i := range audioFormats {
							f := &audioFormats[i]

							lang := "Original / Unknown"
							if f.AudioTrack != nil && f.AudioTrack.DisplayName != "" {
								lang = f.AudioTrack.DisplayName
							}

							title := fmt.Sprintf("%s • %s", lang, f.AudioQuality)
							if f.AudioTrack != nil && f.AudioTrack.AudioIsDefault {
								title += " (Default)"
							}

							items = append(items, formatItem{
								format: f,
								title:  title,
								desc:   fmt.Sprintf("%s • %.2f MB", f.MimeType, float64(f.ContentLength)/1024/1024),
							})
						}
						m.list = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-8)
						m.list.Title = "Select Audio Track"
						return m, nil
					} else {
						// Combined format
						m.state = stateDownloading
						m.fileName = m.video.Title
						go m.runDownload()
						return m, nil
					}
				}
			} else if m.state == stateSelectingAudio {
				item, ok := m.list.SelectedItem().(formatItem)
				if ok {
					m.selectedAudio = item.format
					m.state = stateDownloading
					m.fileName = m.video.Title
					go m.runDownload()
					return m, nil
				}
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case videoInfoFetchedMsg:
		m.video = msg
		m.state = stateSelectingVideo

		var items []list.Item
		for i := range m.video.Formats {
			f := &m.video.Formats[i]

			typeStr := "Video+Audio"
			if f.AudioChannels == 0 && f.MimeType != "" && strings.Contains(f.MimeType, "video") {
				typeStr = "HD (Needs Merge)"
			} else if f.QualityLabel == "" && strings.Contains(f.MimeType, "audio") {
				continue // Don't show audio only in video list
			}

			quality := f.QualityLabel
			if quality == "" {
				quality = f.AudioQuality
			}

			items = append(items, formatItem{
				format: f,
				title:  fmt.Sprintf("%s • %s", quality, typeStr),
				desc:   fmt.Sprintf("%s • %.2f MB", f.MimeType, float64(f.ContentLength)/1024/1024),
			})
		}

		m.list = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-8)
		m.list.Title = "Select Video Quality"
		return m, nil

	case errMsg:
		m.err = msg
		m.state = stateError
		return m, nil

	case downloadProgressMsg:
		var cmd tea.Cmd
		cmd = m.progress.SetPercent(float64(msg))
		return m, cmd

	case mergingMsg:
		m.state = stateMerging
		return m, nil

	case downloadDoneMsg:
		m.fileName = string(msg)
		m.state = stateFinished
		return m, tea.Batch(
			tea.Printf("\n  %s %s\n", statusStyle.Render("Saved as:"), m.fileName),
			tea.Quit,
		)

	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if m2, ok := newModel.(progress.Model); ok {
			m.progress = m2
		}
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.state == stateSelectingVideo || m.state == stateSelectingAudio {
			m.list.SetSize(msg.Width-4, msg.Height-8)
		}
		m.progress.Width = msg.Width - 4
	}

	if m.state == stateSelectingVideo || m.state == stateSelectingAudio {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "\n  Bye! 👋\n\n"
	}

	var s string

	switch m.state {
	case stateFetching:
		s = fmt.Sprintf("\n  %s Fetching video details...\n", m.spinner.View())
	case stateSelectingVideo, stateSelectingAudio:
		s = docStyle.Render(m.list.View())
	case stateDownloading:
		s = fmt.Sprintf("\n  %s\n\n  %s\n\n  %s",
			titleStyle.Render("Downloading Content..."),
			m.progress.View(),
			helpStyle.Render("Downloading multiple streams for HD..."),
		)
	case stateMerging:
		s = fmt.Sprintf("\n  %s %s\n\n  %s",
			m.spinner.View(),
			titleStyle.Render("Merging Tracks..."),
			helpStyle.Render("Using FFmpeg to mux audio and video..."),
		)
	case stateFinished:
		s = fmt.Sprintf("\n  %s\n", titleStyle.Render("Success! Download Complete."))
	case stateError:
		s = fmt.Sprintf("\n  %s\n\n  %v\n",
			errorStyle.Render("Error"),
			m.err,
		)
	}

	return s
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gotube <youtube-url>")
		os.Exit(1)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	p := progress.New(progress.WithDefaultGradient())

	m := &model{
		state:    stateFetching,
		videoURL: os.Args[1],
		spinner:  s,
		progress: p,
	}

	program := tea.NewProgram(m)
	m.program = program

	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
