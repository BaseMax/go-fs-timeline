package timeline

import (
	"fmt"
	"strings"

	"github.com/BaseMax/go-fs-timeline/pkg/database"
	"github.com/fatih/color"
)

type Renderer struct {
	colorEnabled bool
}

func NewRenderer(colorEnabled bool) *Renderer {
	return &Renderer{colorEnabled: colorEnabled}
}

func (r *Renderer) Render(events []*database.Event) string {
	if len(events) == 0 {
		return "No events found.\n"
	}

	var builder strings.Builder

	builder.WriteString(r.header())
	builder.WriteString("\n")

	var currentDate string
	for _, event := range events {
		eventDate := event.Timestamp.Format("2006-01-02")
		if eventDate != currentDate {
			currentDate = eventDate
			builder.WriteString(r.dateHeader(eventDate))
			builder.WriteString("\n")
		}

		builder.WriteString(r.formatEvent(event))
		builder.WriteString("\n")
	}

	builder.WriteString(r.footer(len(events)))

	return builder.String()
}

func (r *Renderer) header() string {
	if r.colorEnabled {
		return color.New(color.FgCyan, color.Bold).Sprint("═══ File System Timeline ═══")
	}
	return "=== File System Timeline ==="
}

func (r *Renderer) dateHeader(date string) string {
	if r.colorEnabled {
		return "  " + color.New(color.FgYellow, color.Bold).Sprintf("▸ %s", date)
	}
	return fmt.Sprintf("  > %s", date)
}

func (r *Renderer) formatEvent(event *database.Event) string {
	timestamp := event.Timestamp.Format("15:04:05")
	eventType := r.colorizeEventType(event.EventType)
	filePath := event.FilePath
	fileType := event.FileType

	return fmt.Sprintf("    %s  %s  %s [%s]", timestamp, eventType, filePath, fileType)
}

func (r *Renderer) colorizeEventType(eventType string) string {
	if !r.colorEnabled {
		return fmt.Sprintf("%-7s", eventType)
	}

	var c *color.Color
	switch eventType {
	case "CREATE":
		c = color.New(color.FgGreen)
	case "WRITE":
		c = color.New(color.FgBlue)
	case "REMOVE":
		c = color.New(color.FgRed)
	case "RENAME":
		c = color.New(color.FgMagenta)
	case "CHMOD":
		c = color.New(color.FgYellow)
	default:
		c = color.New(color.FgWhite)
	}

	return c.Sprintf("%-7s", eventType)
}

func (r *Renderer) footer(count int) string {
	footer := fmt.Sprintf("\nTotal events: %d\n", count)
	if r.colorEnabled {
		return color.New(color.FgCyan).Sprint(footer)
	}
	return footer
}
