package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BaseMax/go-fs-timeline/pkg/database"
	"github.com/BaseMax/go-fs-timeline/pkg/watcher"
	"github.com/spf13/cobra"
)

var (
	watchPath         string
	watchDBPath       string
	watchFlushSeconds int
	watchBufferSize   int
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch a directory for file system changes",
	Long:  `Start monitoring a directory for file system changes and persist events to the database.`,
	RunE:  runWatch,
}

func init() {
	watchCmd.Flags().StringVarP(&watchPath, "path", "p", ".", "Path to watch")
	watchCmd.Flags().StringVarP(&watchDBPath, "db", "d", "fstimeline.db", "Database path")
	watchCmd.Flags().IntVarP(&watchFlushSeconds, "flush", "f", 5, "Flush interval in seconds")
	watchCmd.Flags().IntVarP(&watchBufferSize, "buffer", "b", 100, "Maximum buffer size before flush")
}

func runWatch(cmd *cobra.Command, args []string) error {
	// Open database
	db, err := database.New(watchDBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create watcher
	w, err := watcher.New(db, time.Duration(watchFlushSeconds)*time.Second, watchBufferSize)
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer w.Close()

	// Add path to watch
	if err := w.AddPath(watchPath); err != nil {
		return fmt.Errorf("failed to add path to watcher: %w", err)
	}

	fmt.Printf("👀 Watching directory: %s\n", watchPath)
	fmt.Printf("💾 Database: %s\n", watchDBPath)
	fmt.Printf("⏱️  Flush interval: %d seconds\n", watchFlushSeconds)
	fmt.Printf("📦 Buffer size: %d events\n", watchBufferSize)
	fmt.Println("Press Ctrl+C to stop...")
	fmt.Println()

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down gracefully...")
		cancel()
	}()

	// Start watching
	if err := w.Watch(ctx); err != nil {
		return fmt.Errorf("watcher error: %w", err)
	}

	fmt.Println("✅ Shutdown complete")
	return nil
}
