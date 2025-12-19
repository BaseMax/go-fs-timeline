package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Event struct {
	ID        int64
	Timestamp time.Time
	EventType string
	FilePath  string
	FileName  string
	FileType  string
	Directory string
}

type DB struct {
	conn *sql.DB
}

func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.createSchema(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) createSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		event_type TEXT NOT NULL,
		file_path TEXT NOT NULL,
		file_name TEXT NOT NULL,
		file_type TEXT NOT NULL,
		directory TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_directory ON events(directory);
	CREATE INDEX IF NOT EXISTS idx_file_type ON events(file_type);
	`

	_, err := db.conn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

func (db *DB) InsertEvent(event *Event) error {
	query := `INSERT INTO events (timestamp, event_type, file_path, file_name, file_type, directory)
	          VALUES (?, ?, ?, ?, ?, ?)`

	_, err := db.conn.Exec(query, event.Timestamp, event.EventType, event.FilePath,
		event.FileName, event.FileType, event.Directory)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

func (db *DB) InsertEvents(events []*Event) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO events (timestamp, event_type, file_path, file_name, file_type, directory)
	                         VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, event := range events {
		_, err := stmt.Exec(event.Timestamp, event.EventType, event.FilePath,
			event.FileName, event.FileType, event.Directory)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

type QueryFilter struct {
	StartTime *time.Time
	EndTime   *time.Time
	FileType  string
	Directory string
	Limit     int
}

func (db *DB) QueryEvents(filter QueryFilter) ([]*Event, error) {
	query := `SELECT id, timestamp, event_type, file_path, file_name, file_type, directory
	          FROM events WHERE 1=1`
	args := []interface{}{}

	if filter.StartTime != nil {
		query += " AND timestamp >= ?"
		args = append(args, filter.StartTime)
	}

	if filter.EndTime != nil {
		query += " AND timestamp <= ?"
		args = append(args, filter.EndTime)
	}

	if filter.FileType != "" {
		query += " AND file_type = ?"
		args = append(args, filter.FileType)
	}

	if filter.Directory != "" {
		query += " AND directory LIKE ?"
		args = append(args, filter.Directory+"%")
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		event := &Event{}
		err := rows.Scan(&event.ID, &event.Timestamp, &event.EventType, &event.FilePath,
			&event.FileName, &event.FileType, &event.Directory)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}
