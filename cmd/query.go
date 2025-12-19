package cmd

import (
	"fmt"
	"time"

	"github.com/BaseMax/go-fs-timeline/pkg/database"
	"github.com/BaseMax/go-fs-timeline/pkg/timeline"
	"github.com/spf13/cobra"
)

var (
	queryDBPath   string
	queryStart    string
	queryEnd      string
	queryFileType string
	queryDir      string
	queryLimit    int
	queryNoColor  bool
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query file system events",
	Long:  `Query and display file system events with various filters.`,
	RunE:  runQuery,
}

func init() {
	queryCmd.Flags().StringVarP(&queryDBPath, "db", "d", "fstimeline.db", "Database path")
	queryCmd.Flags().StringVarP(&queryStart, "start", "s", "", "Start time (RFC3339 format or relative like -24h)")
	queryCmd.Flags().StringVarP(&queryEnd, "end", "e", "", "End time (RFC3339 format)")
	queryCmd.Flags().StringVarP(&queryFileType, "type", "t", "", "Filter by file type (e.g., 'go', 'txt')")
	queryCmd.Flags().StringVarP(&queryDir, "dir", "D", "", "Filter by directory")
	queryCmd.Flags().IntVarP(&queryLimit, "limit", "l", 100, "Limit number of results")
	queryCmd.Flags().BoolVarP(&queryNoColor, "no-color", "n", false, "Disable colored output")
}

func runQuery(cmd *cobra.Command, args []string) error {
	// Open database
	db, err := database.New(queryDBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Parse time filters
	filter := database.QueryFilter{
		FileType:  queryFileType,
		Directory: queryDir,
		Limit:     queryLimit,
	}

	if queryStart != "" {
		startTime, err := parseTime(queryStart)
		if err != nil {
			return fmt.Errorf("invalid start time: %w", err)
		}
		filter.StartTime = &startTime
	}

	if queryEnd != "" {
		endTime, err := time.Parse(time.RFC3339, queryEnd)
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

	// Render timeline
	renderer := timeline.NewRenderer(!queryNoColor)
	output := renderer.Render(events)
	fmt.Print(output)

	return nil
}

func parseTime(timeStr string) (time.Time, error) {
	// Try parsing as RFC3339
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t, nil
	}

	// Try parsing as duration (e.g., "-24h", "-1h30m")
	if len(timeStr) > 0 && timeStr[0] == '-' {
		duration, err := time.ParseDuration(timeStr[1:])
		if err == nil {
			return time.Now().Add(-duration), nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format (use RFC3339 or relative like -24h)")
}
