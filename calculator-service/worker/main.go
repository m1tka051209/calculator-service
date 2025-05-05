package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/m1tka051209/calculator-service/config"
	_ "modernc.org/sqlite"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Настройка логгера
	logFile, err := os.OpenFile(cfg.WorkerLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// Подключение к БД
	db, err := sql.Open("sqlite", cfg.WorkerDBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Проверка соединения с БД
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Обработка сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Главный цикл воркера
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	log.Println("Worker started successfully")
	log.Printf("Using DB path: %s", cfg.WorkerDBPath)
	log.Printf("Using log path: %s", cfg.WorkerLogPath)

	for {
		select {
		case <-sigChan:
			log.Println("Received termination signal, shutting down...")
			return
		case <-ticker.C:
			checkTasks(db)
		}
	}
}

func checkTasks(db *sql.DB) {
	// Начинаем транзакцию
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}
	defer tx.Rollback() // В случае ошибки откатываем

	var taskID string
	var expression string
	var status string
	err = tx.QueryRow(`
		SELECT id, expression, status 
		FROM expressions 
		WHERE status = 'pending' 
		ORDER BY created_at ASC 
		LIMIT 1
	`).Scan(&taskID, &expression, &status)
	
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No tasks available (retrying in 1s)")
			return
		}
		log.Printf("Failed to query tasks: %v", err)
		return
	}

	_, err = tx.Exec(`
		UPDATE expressions 
		SET status = 'processing', 
			started_at = CURRENT_TIMESTAMP 
		WHERE id = ?
	`, taskID)
	if err != nil {
		log.Printf("Failed to update task status: %v", err)
		return
	}

	// Фиксируем изменения статуса
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	log.Printf("Processing task %s: %s", taskID, expression)

	result, err := evaluateExpression(expression)
	if err != nil {
		log.Printf("Failed to evaluate expression: %v", err)
		// Обновляем статус на "failed"
		if _, execErr := db.Exec(`
			UPDATE expressions 
			SET status = 'failed', 
				error = ?,
				completed_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, err.Error(), taskID); execErr != nil {
			log.Printf("Failed to update failed status: %v", execErr)
		}
		return
	}

	if _, err := db.Exec(`
		UPDATE expressions 
		SET status = 'completed', 
			result = ?,
			completed_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, result, taskID); err != nil {
		log.Printf("Failed to save result: %v", err)
		return
	}

	log.Printf("Task %s completed with result: %f", taskID, result)
}

func evaluateExpression(expr string) (float64, error) {
	// Удаляем все пробелы из выражения
	expr = strings.ReplaceAll(expr, " ", "")
	if expr == "" {
		return 0, fmt.Errorf("empty expression")
	}

	// Преобразуем строку в токены (числа и операторы)
	tokens, err := tokenize(expr)
	if err != nil {
		return 0, err
	}

	// Преобразуем в обратную польскую нотацию (RPN)
	rpn, err := shuntingYard(tokens)
	if err != nil {
		return 0, err
	}

	// Вычисляем выражение в RPN
	result, err := evaluateRPN(rpn)
	if err != nil {
		return 0, err
	}

	return result, nil
}

// tokenize разбивает строку на токены
func tokenize(expr string) ([]string, error) {
	var tokens []string
	var numBuffer strings.Builder

	for _, ch := range expr {
		if isDigit(ch) || ch == '.' {
			numBuffer.WriteRune(ch)
		} else {
			if numBuffer.Len() > 0 {
				tokens = append(tokens, numBuffer.String())
				numBuffer.Reset()
			}
			if isOperator(ch) {
				tokens = append(tokens, string(ch))
			} else {
				return nil, fmt.Errorf("invalid character: %c", ch)
			}
		}
	}

	if numBuffer.Len() > 0 {
		tokens = append(tokens, numBuffer.String())
	}

	return tokens, nil
}

// shuntingYard преобразует в обратную польскую нотацию
func shuntingYard(tokens []string) ([]string, error) {
	var output []string
	var stack []string

	for _, token := range tokens {
		if isNumber(token) {
			output = append(output, token)
		} else if isOperatorStr(token) {
			for len(stack) > 0 && isOperatorStr(stack[len(stack)-1]) &&
				precedence(stack[len(stack)-1]) >= precedence(token) {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		} else {
			return nil, fmt.Errorf("invalid token: %s", token)
		}
	}

	for len(stack) > 0 {
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

// evaluateRPN вычисляет выражение в RPN
func evaluateRPN(rpn []string) (float64, error) {
	var stack []float64

	for _, token := range rpn {
		if isNumber(token) {
			num, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return 0, err
			}
			stack = append(stack, num)
		} else {
			if len(stack) < 2 {
				return 0, fmt.Errorf("invalid expression")
			}
			a, b := stack[len(stack)-2], stack[len(stack)-1]
			stack = stack[:len(stack)-2]

			var result float64
			switch token {
			case "+":
				result = a + b
			case "-":
				result = a - b
			case "*":
				result = a * b
			case "/":
				if b == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				result = a / b
			default:
				return 0, fmt.Errorf("unknown operator: %s", token)
			}
			stack = append(stack, result)
		}
	}

	if len(stack) != 1 {
		return 0, fmt.Errorf("invalid expression")
	}

	return stack[0], nil
}

// Вспомогательные функции
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isOperator(ch rune) bool {
	return ch == '+' || ch == '-' || ch == '*' || ch == '/'
}

func isOperatorStr(s string) bool {
	return s == "+" || s == "-" || s == "*" || s == "/"
}

func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}