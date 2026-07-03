package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go-worker/internal/domain"
)

type StatusProvider interface {
	FetchStatus(ctx context.Context, orderID domain.OrderID) (domain.OrderProcessingResult, error)
}

type StatusChecker struct {
	provider       StatusProvider
	requestTimeout time.Duration
}

func NewStatusChecker(provider StatusProvider, timeout time.Duration) *StatusChecker {
	return &StatusChecker{
		provider:       provider,
		requestTimeout: timeout,
	}
}

func (c *StatusChecker) CheckOrders(ctx context.Context, orderIDs []string) ([]domain.OrderProcessingResult, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan domain.OrderProcessingResult, len(orderIDs))
	var wg sync.WaitGroup

	for _, id := range orderIDs {
		wg.Add(1)
		go func(orderID string) {
			defer wg.Done()

			requestCtx, requestCancel := context.WithTimeout(ctx, c.requestTimeout)
			defer requestCancel()

			result, err := c.provider.FetchStatus(requestCtx, domain.OrderID(orderID))
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					err = NewOrderTimeoutError(domain.OrderID(orderID))
				}
				result = domain.OrderProcessingResult{
					ID:       domain.OrderID(orderID),
					Status:   "",
					Duration: result.Duration,
					Err:      err,
				}
			}

			results <- result
		}(id)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var collected []domain.OrderProcessingResult
	for result := range results {
		collected = append(collected, result)
	}

	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return collected, fmt.Errorf("fluxo de verificação interrompido: %w", err)
	}

	return collected, nil
}

type OrderProcessingError struct {
	OrderID domain.OrderID
	Reason  string
}

func (e OrderProcessingError) Error() string {
	return fmt.Sprintf("pedido %s: %s", e.OrderID, e.Reason)
}

func NewOrderTimeoutError(orderID domain.OrderID) error {
	return OrderProcessingError{
		OrderID: orderID,
		Reason:  "timeout na solicitação ao provedor externo",
	}
}
