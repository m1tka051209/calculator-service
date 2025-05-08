package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/m1tka051209/calculator-service/models"
	_ "modernc.org/sqlite"
)

type Repository interface {
	CreateUser(ctx context.Context, login, passwordHash string) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	CreateExpression(ctx context.Context, userID, expr string) (string, error)
	GetExpressionsByUser(ctx context.Context, userID string) ([]models.Expression, error)
	GetPendingTasks(ctx context.Context, limit int) ([]models.Task, error)
	UpdateTaskResult(ctx context.Context, taskID string, result float64) error
	UpdateTaskStatus(ctx context.Context, taskID, status string) error
}

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_timeout=5000&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &SQLiteRepository{db: db}, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			login TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS expressions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			expression TEXT NOT NULL,
			status TEXT NOT NULL,
			result REAL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);
		
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			expression_id TEXT NOT NULL,
			arg1 REAL NOT NULL,
			arg2 REAL NOT NULL,
			operation TEXT NOT NULL,
			operation_time INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			result REAL,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			FOREIGN KEY(expression_id) REFERENCES expressions(id)
		);
	`)
	return err
}

func (r *SQLiteRepository) CreateUser(ctx context.Context, login, passwordHash string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO users(id, login, password_hash) VALUES(?, ?, ?)",
		uuid.New().String(), login, passwordHash)
	return err
}

func (r *SQLiteRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, login, password_hash FROM users WHERE login = ?", login).
		Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SQLiteRepository) CreateExpression(ctx context.Context, userID, expr string) (string, error) {
	id := uuid.New().String()
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO expressions(id, user_id, expression, status) VALUES(?, ?, ?, 'pending')",
		id, userID, expr)
	return id, err
}

func (r *SQLiteRepository) GetExpressionsByUser(ctx context.Context, userID string) ([]models.Expression, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT id, expression, status, result, created_at, started_at, completed_at 
         FROM expressions WHERE user_id = ? ORDER BY created_at DESC`, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var exprs []models.Expression
    for rows.Next() {
        var e models.Expression
        var startedAt, completedAt sql.NullTime
        
        err := rows.Scan(
            &e.ID,
            &e.Expression,
            &e.Status,
            &e.Result,
            &e.CreatedAt,
            &startedAt,
            &completedAt,
        )
        if err != nil {
            return nil, err
        }

        if startedAt.Valid {
            e.StartedAt = &startedAt.Time
        }
        if completedAt.Valid {
            e.CompletedAt = &completedAt.Time
        }

        exprs = append(exprs, e)
    }
    return exprs, nil
}

func (r *SQLiteRepository) GetPendingTasks(ctx context.Context, limit int) ([]models.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx,
		`SELECT id, expression_id, arg1, arg2, operation, operation_time 
		 FROM tasks WHERE status = 'pending' LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		err := rows.Scan(&t.ID, &t.ExpressionID, &t.Arg1, &t.Arg2, &t.Operation, &t.OperationTime)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *SQLiteRepository) UpdateTaskResult(ctx context.Context, taskID string, result float64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE tasks SET 
			status = 'completed', 
			result = ?,
			completed_at = CURRENT_TIMESTAMP 
		 WHERE id = ?`,
		result, taskID)
	return err
}

func (r *SQLiteRepository) UpdateTaskStatus(ctx context.Context, taskID, status string) error {
	query := "UPDATE tasks SET status = ?"
	args := []interface{}{status}

	switch status {
	case "processing":
		query += ", started_at = CURRENT_TIMESTAMP"
	case "completed":
		query += ", completed_at = CURRENT_TIMESTAMP"
	case "failed":
		query += ", completed_at = CURRENT_TIMESTAMP"
	}

	query += " WHERE id = ?"
	args = append(args, taskID)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}