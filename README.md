# 🧮 Distributed Arithmetic Calculator | Калькулятор арифметических выражений

![Go Version](https://img.shields.io/badge/go-1.21%2B-blue)
![Database](https://img.shields.io/badge/database-SQLite-lightgrey)

## 📝 Описание проекта

Распределенный сервис для вычисления арифметических выражений с:
- Многопользовательской аутентификацией (JWT)
- Асинхронной обработкой задач
- Устойчивым хранением в SQLite
- Внутренним gRPC взаимодействием
- REST API для клиентов

## 🚀 Быстрый старт

### Требования
- Go 1.21+
- SQLite3
- Git

### Установка и запуск
```bash
# 1. Клонирование репозитория
git clone https://github.com/m1tka051209/calculator-service.git
cd calculator-service

# 2. Установка зависимостей
go mod download

# 3. Запуск сервиса
go run main.go

🔧 Функционал
Регистрация пользователя
bash
curl --location 'http://localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login": "user1",
  "password": "password123"
}'
Авторизация
bash
curl --location 'http://localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "login": "user1",
  "password": "password123"
}'
Ответ:

json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
Создание выражения
bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer YOUR_JWT_TOKEN' \
--data '{
  "expression": "2+2*2"
}'
Ответ:

json
{
  "expression_id": "generated-id-123",
  "status": "pending"
}
Получение списка выражений
bash
curl --location 'http://localhost:8080/api/v1/expressions' \
--header 'Authorization: Bearer YOUR_JWT_TOKEN'
Ответ:

json
[
  {
    "id": "generated-id-123",
    "expression": "2+2*2",
    "status": "completed",
    "result": 6,
    "created_at": "2025-05-09T16:20:00Z",
    "completed_at": "2025-05-09T16:20:05Z"
  }
]

📊 База данных
Используется SQLite с тремя основными таблицами:

users - хранение пользователей

expressions - хранение выражений

tasks - хранение задач для вычислений

🔒 Аутентификация
Используется JWT (JSON Web Tokens) для аутентификации пользователей.
Токен должен передаваться в заголовке Authorization: Bearer YOUR_TOKEN.
