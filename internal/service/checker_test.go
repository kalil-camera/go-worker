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
		err:    errors.New("falha simulada do provedor"),
		delay:  100 * time.Millisecond,
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
func TestCheckOrders_MultipleOrders(t *testing.T) {
	provider := &fakeProvider{
		result: domain.OrderProcessingResult{Status: "Processando", Duration: 50 * time.Millisecond},
		err:    nil,
		delay:  10 * time.Millisecond,
	}
	checker := NewStatusChecker(provider, 1*time.Second)
	orderIDs := []string{"ORD-0004", "ORD-0005", "ORD-0006"}

	results, err := checker.CheckOrders(context.Background(), orderIDs)
	if err != nil {
		t.Fatalf("não esperava erro, obteve: %v", err)
	}
	if len(results) != len(orderIDs) {
		t.Fatalf("esperava %d resultados, obteve %d", len(orderIDs), len(results))
	}

	seen := make(map[string]bool, len(orderIDs))
	for _, result := range results {
		if result.Err != nil {
			t.Fatalf("não esperava erro para %s, obteve: %v", result.ID, result.Err)
		}
		seen[string(result.ID)] = true
	}
	for _, id := range orderIDs {
		if !seen[id] {
			t.Fatalf("resultado ausente para %s", id)
		}
	}
}

func TestCheckOrders_ParentContextCancel(t *testing.T) {
	parentCtx, cancel := context.WithCancel(context.Background())
	provider := &fakeProvider{
		result: domain.OrderProcessingResult{Status: "Processando", Duration: 500 * time.Millisecond},
		err:    nil,
		delay:  500 * time.Millisecond,
	}
	checker := NewStatusChecker(provider, 1*time.Second)

	go func() {
		time.Sleep(25 * time.Millisecond)
		cancel()
	}()

	results, err := checker.CheckOrders(parentCtx, []string{"ORD-0007"})
	if err != nil {
		t.Fatalf("não esperava erro de fluxo principal, obteve: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("esperava 1 resultado, obteve %d", len(results))
	}
	if results[0].Err == nil {
		t.Fatal("esperava erro de cancelamento, obteve nil")
	}
	if !errors.Is(results[0].Err, context.Canceled) {
		t.Fatalf("esperava erro context.Canceled, obteve: %v", results[0].Err)
	}
}

func TestOrderProcessingError_Error(t *testing.T) {
	err := NewOrderTimeoutError(domain.OrderID("ORD-0008"))
	if got := err.Error(); got != "pedido ORD-0008: timeout na solicitação ao provedor externo" {
		t.Fatalf("mensagem inesperada: %q", got)
	}
}
