package domain

import "time"

type OrderID string

type OrderProcessingResult struct {
    ID       OrderID
    Status   string
    Duration time.Duration
    Err      error
}
