package db

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    "github.com/google/uuid"
    "github.com/mattn/go-sqlite3"
    "github.com/m1tka051209/calculator-service/models"
)

type Repository struct {
    db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*Repository, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    if err := createTables(db); err != nil {
        return nil, fmt.Errorf("failed to create tables: %w", err)
    }

    return &Repository{db: db}, nil
}

func createTables(db *sql.DB) error {
    _, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        login TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE TABLE IF NOT EXISTS expressions (
        id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL,
        expression TEXT NOT NULL,
        status TEXT NOT NULL,
        result REAL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
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
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(expression_id) REFERENCES expressions(id)
    );`)
    return err
}

func (r *Repository) CreateUser(ctx context.Context, login, passwordHash string) error {
    _, err := r.db.ExecContext(ctx,
        "INSERT INTO users(id, login, password_hash) VALUES(?, ?, ?)",
        uuid.New().String(), login, passwordHash)
    return err
}

func (r *Repository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
    var user models.User
    err := r.db.QueryRowContext(ctx,
        "SELECT id, login, password_hash FROM users WHERE login = ?", login).
        Scan(&user.ID, &user.Login, &user.PasswordHash)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *Repository) GetPendingTasks(ctx context.Context, limit int) ([]models.Task, error) {
    rows, err := r.db.QueryContext(ctx,
        `SELECT id, expression_id, arg1, arg2, operation, operation_time 
         FROM tasks WHERE status = 'pending' LIMIT ?`, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tasks []models.Task
    for rows.Next() {
        var t models.Task
        if err := rows.Scan(&t.ID, &t.ExpressionID, &t.Arg1, &t.Arg2, &t.Operation, &t.OperationTime); err != nil {
            return nil, err
        }
        tasks = append(tasks, t)
    }
    return tasks, nil
}

func (r *Repository) UpdateTaskResult(ctx context.Context, taskID string, result float64) error {
    _, err := r.db.ExecContext(ctx,
        `UPDATE tasks SET status = 'completed', result = ? WHERE id = ?`,
        result, taskID)
    return err
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, login, password_hash FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}