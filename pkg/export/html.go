package export

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/BaseMax/go-fs-timeline/pkg/database"
)

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>File System Timeline</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 10px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 2.5em;
        }
        .header p {
            margin: 10px 0 0 0;
            opacity: 0.9;
        }
        .timeline {
            padding: 30px;
        }
        .date-group {
            margin-bottom: 30px;
        }
        .date-header {
            font-size: 1.3em;
            color: #667eea;
            font-weight: bold;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 2px solid #667eea;
        }
        .event {
            display: flex;
            align-items: center;
            padding: 15px;
            margin: 10px 0;
            background: #f8f9fa;
            border-left: 4px solid #ddd;
            border-radius: 5px;
            transition: all 0.3s;
        }
        .event:hover {
            transform: translateX(5px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.1);
        }
        .event-time {
            font-weight: bold;
            color: #666;
            min-width: 80px;
        }
        .event-type {
            padding: 5px 10px;
            border-radius: 4px;
            font-weight: bold;
            min-width: 80px;
            text-align: center;
            margin: 0 15px;
        }
        .event-type-CREATE { background: #d4edda; color: #155724; border-left-color: #28a745; }
        .event-type-WRITE { background: #cce5ff; color: #004085; border-left-color: #007bff; }
        .event-type-REMOVE { background: #f8d7da; color: #721c24; border-left-color: #dc3545; }
        .event-type-RENAME { background: #e2d5f0; color: #5a2d7a; border-left-color: #9b59b6; }
        .event-type-CHMOD { background: #fff3cd; color: #856404; border-left-color: #ffc107; }
        .event-path {
            flex: 1;
            color: #333;
            word-break: break-all;
        }
        .event-filetype {
            background: #e9ecef;
            padding: 5px 10px;
            border-radius: 4px;
            font-size: 0.9em;
            color: #495057;
        }
        .footer {
            background: #f8f9fa;
            padding: 20px;
            text-align: center;
            color: #666;
            border-top: 1px solid #dee2e6;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📂 File System Timeline</h1>
            <p>Generated on {{.GeneratedAt}}</p>
            <p>Total Events: {{.TotalEvents}}</p>
        </div>
        <div class="timeline">
            {{range $date, $events := .EventsByDate}}
            <div class="date-group">
                <div class="date-header">📅 {{$date}}</div>
                {{range $events}}
                <div class="event">
                    <div class="event-time">{{.TimeStr}}</div>
                    <div class="event-type event-type-{{.EventType}}">{{.EventType}}</div>
                    <div class="event-path">{{.FilePath}}</div>
                    <div class="event-filetype">{{.FileType}}</div>
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>File System Timeline Monitor - github.com/BaseMax/go-fs-timeline</p>
        </div>
    </div>
</body>
</html>`

type HTMLExporter struct {
	tmpl *template.Template
}

type eventData struct {
	TimeStr   string
	EventType string
	FilePath  string
	FileType  string
}

type templateData struct {
	GeneratedAt  string
	TotalEvents  int
	EventsByDate map[string][]eventData
}

func NewHTMLExporter() (*HTMLExporter, error) {
	tmpl, err := template.New("timeline").Parse(htmlTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &HTMLExporter{tmpl: tmpl}, nil
}

func (e *HTMLExporter) Export(events []*database.Event, outputPath string) error {
	// Group events by date
	eventsByDate := make(map[string][]eventData)
	for _, event := range events {
		dateStr := event.Timestamp.Format("2006-01-02")
		timeStr := event.Timestamp.Format("15:04:05")

		eventsByDate[dateStr] = append(eventsByDate[dateStr], eventData{
			TimeStr:   timeStr,
			EventType: event.EventType,
			FilePath:  event.FilePath,
			FileType:  event.FileType,
		})
	}

	data := templateData{
		GeneratedAt:  time.Now().Format("2006-01-02 15:04:05"),
		TotalEvents:  len(events),
		EventsByDate: eventsByDate,
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	if err := e.tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
