package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(false)
)

// ProgressBar is a simple progress bar writer
type ProgressBar struct {
	total    int64
	current  int64
	width    int
	label    string
	finished bool
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int64, label string) *ProgressBar {
	return &ProgressBar{
		total: total,
		width: 50,
		label: label,
	}
}

// Write implements io.Writer to track bytes written
func (p *ProgressBar) Write(b []byte) (int, error) {
	n := len(b)
	p.current += int64(n)
	p.render()
	return n, nil
}

// SetCurrent sets the current progress value
func (p *ProgressBar) SetCurrent(current int64) {
	p.current = current
	p.render()
}

// Finish marks the progress bar as complete
func (p *ProgressBar) Finish() {
	p.finished = true
	p.render()
	fmt.Println() // New line after progress bar
}

// render renders the progress bar
func (p *ProgressBar) render() {
	if p.total == 0 {
		// Indeterminate progress
		fmt.Printf("\r%s %s", 
			infoStyle.Render(p.label),
			infoStyle.Render("..."))
		return
	}

	percent := float64(p.current) / float64(p.total)
	if percent > 1.0 {
		percent = 1.0
	}

	filled := int(float64(p.width) * percent)
	empty := p.width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	
	// Format bytes
	var currentStr, totalStr string
	if p.total > 1024*1024 {
		currentStr = fmt.Sprintf("%.1f MB", float64(p.current)/(1024*1024))
		totalStr = fmt.Sprintf("%.1f MB", float64(p.total)/(1024*1024))
	} else if p.total > 1024 {
		currentStr = fmt.Sprintf("%.1f KB", float64(p.current)/1024)
		totalStr = fmt.Sprintf("%.1f KB", float64(p.total)/1024)
	} else {
		currentStr = fmt.Sprintf("%d B", p.current)
		totalStr = fmt.Sprintf("%d B", p.total)
	}

	progressText := fmt.Sprintf("%s [%s] %s / %s (%.1f%%)",
		infoStyle.Render(p.label),
		lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(bar),
		currentStr,
		totalStr,
		percent*100,
	)

	fmt.Printf("\r%s", progressText)
	os.Stdout.Sync()
}

// FileProgressBar is a simple progress bar for file count
type FileProgressBar struct {
	total   int
	current int
	label   string
}

// NewFileProgressBar creates a new file progress bar
func NewFileProgressBar(total int, label string) *FileProgressBar {
	return &FileProgressBar{
		total: total,
		label: label,
	}
}

// Increment increments the file count
func (p *FileProgressBar) Increment() {
	p.current++
	p.render()
}

// SetCurrent sets the current file count
func (p *FileProgressBar) SetCurrent(current int) {
	p.current = current
	p.render()
}

// Finish marks the progress bar as complete
func (p *FileProgressBar) Finish() {
	p.render()
	fmt.Println() // New line after progress bar
}

// render renders the file progress bar
func (p *FileProgressBar) render() {
	if p.total == 0 {
		// Indeterminate progress - just show count
		progressText := fmt.Sprintf("%s %d files...",
			infoStyle.Render(p.label),
			p.current,
		)
		fmt.Printf("\r%s", progressText)
		os.Stdout.Sync()
		return
	}

	percent := float64(p.current) / float64(p.total)
	if percent > 1.0 {
		percent = 1.0
	}

	width := 30
	filled := int(float64(width) * percent)
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	progressText := fmt.Sprintf("%s [%s] %d / %d files (%.1f%%)",
		infoStyle.Render(p.label),
		lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(bar),
		p.current,
		p.total,
		percent*100,
	)

	fmt.Printf("\r%s", progressText)
	os.Stdout.Sync()
}

// ProgressWriter wraps an io.Writer to track progress
type ProgressWriter struct {
	writer      io.Writer
	progressBar *ProgressBar
}

// NewProgressWriter creates a new progress writer
func NewProgressWriter(writer io.Writer, progressBar *ProgressBar) *ProgressWriter {
	return &ProgressWriter{
		writer:      writer,
		progressBar: progressBar,
	}
}

// Write implements io.Writer
func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if pw.progressBar != nil {
		pw.progressBar.Write(p)
	}
	return n, err
}

