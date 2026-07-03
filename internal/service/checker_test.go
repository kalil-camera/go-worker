package service

import (
    "context"
    "errors"
    "testing"
    "time"

    "go-worker/internal/domain"
)

type fakeProvider struct {
    result domain.OrderProcessingResult
    err    error
    delay  time.Duration
}

func (f *fakeProvider) FetchStatus(ctx context.Context, orderID domain.OrderID) (domain.OrderProcessingResult, error) {
    select {
    case <-time.After(f.delay):
    case <-ctx.Done():
        return domain.OrderProcessingResult{ID: orderID}, ctx.Err()
    }
    f.result.ID = orderID
    return f.result, f.err
}

func TestCheckOrders_Success(t *testing.T) {
    provider := &fakeProvider{
        result: domain.OrderProcessingResult{Status: "Entregue", Duration: 250 * time.Millisecond},
        err:    nil,
        delay:  250 * time.Millisecond,
    }
    checker := NewStatusChecker(provider, 1*time.Second)

    results, err := checker.CheckOrders(context.Background(), []string{"ORD-0001"})
    if err != nil {
        t.Fatalf("esperava sem erro, obteve: %v", err)
    }
    if len(results) != 1 {
        t.Fatalf("esperava 1 resultado, obteve %d", len(results))
    }
    if results[0].Err != nil {
        t.Fatalf("esperava resultado sem erro, obteve: %v", results[0].Err)
    }
    if results[0].Status != "Entregue" {
        t.Fatalf("esperava status Entregue, obteve: %q", results[0].Status)
    }
}

func TestCheckOrders_Timeout(t *testing.T) {
    provider := &fakeProvider{
        result: domain.OrderProcessingResult{Status: "Processando", Duration: 0},
        err:    nil,
        delay:  2 * time.Second,
    }
    checker := NewStatusChecker(provider, 100*time.Millisecond)

    results, err := checker.CheckOrders(context.Background(), []string{"ORD-0002"})
    if err != nil {
        t.Fatalf("não esperava erro de fluxo principal, obteve: %v", err)
    }
    if len(results) != 1 {
        t.Fatalf("esperava 1 resultado, obteve %d", len(results))
    }
    if results[0].Err == nil {
        t.Fatal("esperava erro de timeout, obteve nil")
    }
    var timeoutError OrderProcessingError
    if !errors.As(results[0].Err, &timeoutError) {
        t.Fatalf("esperava tipo OrderProcessingError, obteve: %T", results[0].Err)
    }
}

func TestCheckOrders_ProviderError(t *testing.T) {
    provider := &fakeProvider{
        result: domain.OrderProcessingResult{Status: "", Duration: 100 * time.Millisecond},
        err: &OrderProviderError{
            OrderID: "ORD-0003",
            Reason:  "falha simulada",
        },
        delay: 100 * time.Millisecond,
    }
    checker := NewStatusChecker(provider, 1*time.Second)

    results, err := checker.CheckOrders(context.Background(), []string{"ORD-0003"})
    if err != nil {
        t.Fatalf("não esperava erro de fluxo principal, obteve: %v", err)
    }
    if len(results) != 1 {
        t.Fatalf("esperava 1 resultado, obteve %d", len(results))
    }
    if results[0].Err == nil {
        t.Fatal("esperava erro de provedor, obteve nil")
    }
    if results[0].Status != "" {
        t.Fatalf("esperava status vazio em caso de erro, obteve %q", results[0].Status)
    }
}
