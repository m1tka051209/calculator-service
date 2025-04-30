package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
)

type Task struct {
    ID            string  `json:"id"`
    Arg1          float64 `json:"arg1"`
    Arg2          float64 `json:"arg2"`
    Operation     string  `json:"operation"`
    OperationTime int     `json:"operation_time"`
}

func main() {
    for {
        task, err := fetchTask()
        if err != nil {
            log.Println("Fetch task error:", err)
            time.Sleep(2 * time.Second)
            continue
        }

        result := calculate(task)
        if err := submitResult(task.ID, result); err != nil {
            log.Println("Submit result error:", err)
        }
    }
}

func fetchTask() (*Task, error) {
    resp, err := http.Get("http://localhost:8080/internal/task")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return nil, nil
    }

    var task Task
    if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
        return nil, err
    }
    return &task, nil
}

func submitResult(taskID string, result float64) error {
    payload := struct {
        TaskID string  `json:"task_id"`
        Result float64 `json:"result"`
    }{taskID, result}

    jsonData, _ := json.Marshal(payload)
    resp, err := http.Post("http://localhost:8080/internal/result", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }
    return nil
}