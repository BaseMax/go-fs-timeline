package cmd

import (
	"fmt"
	"time"

	"github.com/BaseMax/go-fs-timeline/pkg/database"
	"github.com/BaseMax/go-fs-timeline/pkg/export"
	"github.com/spf13/cobra"
)

var (
	exportDBPath   string
	exportOutput   string
	exportStart    string
	exportEnd      string
	exportFileType string
	exportDir      string
	exportLimit    int
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export timeline to HTML",
	Long:  `Export file system timeline events to an HTML file.`,
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringVarP(&exportDBPath, "db", "d", "fstimeline.db", "Database path")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "timeline.html", "Output HTML file")
	exportCmd.Flags().StringVarP(&exportStart, "start", "s", "", "Start time (RFC3339 format or relative like -24h)")
	exportCmd.Flags().StringVarP(&exportEnd, "end", "e", "", "End time (RFC3339 format)")
	exportCmd.Flags().StringVarP(&exportFileType, "type", "t", "", "Filter by file type")
	exportCmd.Flags().StringVarP(&exportDir, "dir", "D", "", "Filter by directory")
	exportCmd.Flags().IntVarP(&exportLimit, "limit", "l", 1000, "Limit number of results")
}

func runExport(cmd *cobra.Command, args []string) error {
	// Open database
	db, err := database.New(exportDBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Parse time filters
	filter := database.QueryFilter{
		FileType:  exportFileType,
		Directory: exportDir,
		Limit:     exportLimit,
	}

	if exportStart != "" {
		startTime, err := parseTime(exportStart)
		if err != nil {
			return fmt.Errorf("invalid start time: %w", err)
		}
		filter.StartTime = &startTime
	}

	if exportEnd != "" {
		endTime, err := time.Parse(time.RFC3339, exportEnd)
		if err != nil {
			return fmt.Errorf("invalid end time: %w", err)
		}
		filter.EndTime = &endTime
	}

	// Query events
	events, err := db.QueryEvents(filter)
	if err != nil {
		return fmt.Errorf("failed to query events: %w", err)
	}

	// Reverse to show oldest first
	for i := len(events)/2 - 1; i >= 0; i-- {
		opp := len(events) - 1 - i
		events[i], events[opp] = events[opp], events[i]
	}

	// Create exporter
	exporter, err := export.NewHTMLExporter()
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	// Export to HTML
	if err := exporter.Export(events, exportOutput); err != nil {
		return fmt.Errorf("failed to export: %w", err)
	}

	fmt.Printf("✅ Timeline exported to: %s\n", exportOutput)
	fmt.Printf("📊 Total events: %d\n", len(events))

	return nil
}
