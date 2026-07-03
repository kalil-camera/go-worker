package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "go-worker/internal/repository"
    "go-worker/internal/service"
)

func main() {
    orderIDs := []string{
        "ORD-1001",
        "ORD-1002",
        "ORD-1003",
        "ORD-1004",
        "ORD-1005",
    }

    provider := repository.NewSimulatedProvider(time.Now().UnixNano())
    checker := service.NewStatusChecker(provider, 1200*time.Millisecond)

    ctx := context.Background()
    results, err := checker.CheckOrders(ctx, orderIDs)
    if err != nil {
        log.Printf("Aviso: %v", err)
    }

    fmt.Println("Status de pedidos:")
    for _, result := range results {
        if result.Err != nil {
            fmt.Printf("- %s | ERRO: %v | tempo: %s\n", result.ID, result.Err, result.Duration)
            continue
        }

        fmt.Printf("- %s | %s | tempo: %s\n", result.ID, result.Status, result.Duration)
    }
}
