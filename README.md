# go-fs-timeline

A historical timeline viewer for file system changes. This tool monitors file system changes in real-time using OS-native watchers, persists events to a local SQLite database, and provides powerful querying and visualization capabilities.

## Features

- 🔍 **Real-time Monitoring**: Uses OS-native file system watchers (via fsnotify) for efficient change detection
- 💾 **Persistent Storage**: Events stored in SQLite database with indexed queries
- ⚡ **Low CPU Usage**: Event batching and efficient buffering minimize resource consumption
- 🎨 **Terminal Timeline**: Colorful, organized timeline view in your terminal
- 📊 **HTML Export**: Generate beautiful HTML reports of file system activity
- 🔎 **Flexible Querying**: Filter by time range, file type, or directory
- 🛡️ **Robust**: Graceful shutdown handling and error recovery

## Installation

### From Source

```bash
git clone https://github.com/BaseMax/go-fs-timeline
cd go-fs-timeline
go build -o fstimeline .
```

### Requirements

- Go 1.24 or later
- CGO enabled (required for SQLite)

## Usage

### Watch Mode

Monitor a directory for file system changes:

```bash
# Watch current directory
./fstimeline watch

# Watch specific directory
./fstimeline watch -p /path/to/directory

# Custom database location
./fstimeline watch -p /path/to/dir -d /path/to/timeline.db

# Adjust flush interval and buffer size for performance
./fstimeline watch -f 10 -b 200
```

**Options:**
- `-p, --path`: Path to watch (default: current directory)
- `-d, --db`: Database path (default: fstimeline.db)
- `-f, --flush`: Flush interval in seconds (default: 5)
- `-b, --buffer`: Maximum buffer size before flush (default: 100)

### Query Mode

Query and display historical events:

```bash
# Show recent events
./fstimeline query

# Filter by file type
./fstimeline query -t go
./fstimeline query -t txt

# Filter by directory
./fstimeline query -D /path/to/dir

# Filter by time (last 24 hours)
./fstimeline query -s -24h

# Filter by time range (RFC3339 format)
./fstimeline query -s 2025-12-19T00:00:00Z -e 2025-12-19T23:59:59Z

# Limit results
./fstimeline query -l 50

# Disable colors
./fstimeline query --no-color
```

**Options:**
- `-d, --db`: Database path (default: fstimeline.db)
- `-s, --start`: Start time (RFC3339 or relative like -24h)
- `-e, --end`: End time (RFC3339)
- `-t, --type`: Filter by file type (e.g., 'go', 'txt')
- `-D, --dir`: Filter by directory
- `-l, --limit`: Limit number of results (default: 100)
- `-n, --no-color`: Disable colored output

### Export Mode

Export timeline to HTML:

```bash
# Export all events
./fstimeline export -o timeline.html

# Export with filters
./fstimeline export -t go -s -7d -o go-changes.html

# Export from specific directory
./fstimeline export -D /src -o src-timeline.html
```

**Options:**
- `-d, --db`: Database path (default: fstimeline.db)
- `-o, --output`: Output HTML file (default: timeline.html)
- `-s, --start`: Start time filter
- `-e, --end`: End time filter
- `-t, --type`: Filter by file type
- `-D, --dir`: Filter by directory
- `-l, --limit`: Limit number of results (default: 1000)

## Examples

### Monitor a project directory

```bash
# Start monitoring
./fstimeline watch -p /home/user/projects/myapp

# In another terminal, query recent changes
./fstimeline query -s -1h

# Export daily report
./fstimeline export -s -24h -o daily-report.html
```

### Track Go source files

```bash
# Query only Go files
./fstimeline query -t go -l 50

# Export Go file changes
./fstimeline export -t go -o go-changes.html
```

### Monitor specific subdirectory

```bash
# Query events in specific directory
./fstimeline query -D /home/user/projects/myapp/src

# Export that directory's timeline
./fstimeline export -D /home/user/projects/myapp/src -o src-timeline.html
```

## Architecture

### Components

- **Watcher**: Uses fsnotify for OS-native file system monitoring
- **Database**: SQLite with indexed tables for fast queries
- **Event Buffer**: Batches events to minimize database writes
- **Timeline Renderer**: Colorful terminal output using fatih/color
- **HTML Exporter**: Template-based HTML generation

### Event Types

- `CREATE`: New file or directory created
- `WRITE`: File content modified
- `REMOVE`: File or directory deleted
- `RENAME`: File or directory renamed
- `CHMOD`: File permissions changed

### Performance Characteristics

- **CPU Usage**: Minimal - event-driven architecture with batching
- **Memory**: Low - configurable buffer size (default 100 events)
- **Database**: Indexed queries for fast filtering
- **I/O**: Batch writes reduce disk operations

## Database Schema

```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL,
    event_type TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_type TEXT NOT NULL,
    directory TEXT NOT NULL
);

CREATE INDEX idx_timestamp ON events(timestamp);
CREATE INDEX idx_directory ON events(directory);
CREATE INDEX idx_file_type ON events(file_type);
```

## License

This project is licensed under the GPL-3.0 License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Author

Max Base - [@BaseMax](https://github.com/BaseMax)
