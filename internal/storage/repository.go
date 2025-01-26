package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

//Initialize SQLite database and create tables
func NewSQLiteDB(dbPath string) (*sql.DB, error){
	db, err := sql.Open("sqlite3", dbPath)
	if err!=nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	//Create message table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			user_id INTEGER,
			message TEXT,
			timestamp DATETIME
		);
	
		CREATE TABLE IF NOT EXISTS usage_metrics (
			user_id BIGINT NOT NULL,
			date DATE NOT NULL,
			tokens_used INT DEFAULT 0,
			prompts_count INT DEFAULT 0,
			PRIMARY KEY (user_id, date)
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return db, nil
}

// Save a message to the database
func (r *Repository) SaveMessage(userID int64, message string) error {
	_, err := r.db.Exec(
		"INSERT INTO messages (user_id, message, timestamp) VALUES (?, ?, ?)",
		userID,message, time.Now().UTC(),
	)
	return err
}

// Retrieve the last N messages for a user (for context)
func (r *Repository) GetLastMessages(userID int64, limit int) ([]string, error){
	rows, err := r.db.Query(
		"SELECT message FROM messages WHERE user_id = ? ORDER BY timestamp DESC LIMIT ?",
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []string
	for rows.Next(){
		var msg string
		if err := rows.Scan(&msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	//Reverese to maintain chronological order
	for i, j := 0, len(messages)-1; i<j; i, j =i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *Repository) GetDailyUsage(userID int64, date time.Time)(int, error){
	var tokensUsed int
	queryDate := date.Format("2006-01-02")

	err := r.db.QueryRow(
		"SELECT tokens_used FROM usage_metrics WHERE user_id = ? AND date = ?",
		userID, queryDate,
	).Scan(&tokensUsed)

	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("error getting daily usage: %w", err)
	}

	return tokensUsed, nil
}

func (r *Repository) GetDailyPrompts(userID int64, date time.Time)(int, error) {
	var promptsCount int
	queryDate:=date.Format("2006-01-02")

	err := r.db.QueryRow(
		"SELECT prompts_count FROM usage_metrics WHERE user_id = ? AND date = ?", userID, queryDate,
	).Scan(&promptsCount)

	if err !=nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("error gettign daily prompts: %w", err)
	}

	return promptsCount, nil
}

func (r *Repository) RecordUsage(userID int64, tokens int) error {
	queryDate := time.Now().UTC().Format("2006-01-02")

	_, err := r.db.Exec(`
		INSERT INTO usage_metrics (user_id, date, tokens_used, prompts_count)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(user_id, date)
		DO UPDATE SET
			tokens_used = tokens_used + ?,
			prompts_count = prompts_count + 1
		`, userID, queryDate, tokens, tokens,)

		if err != nil {
			return fmt.Errorf("error recording usage: %w", err)
		}
		return nil
}

