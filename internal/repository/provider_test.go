package repository

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSimulatedProvider_FetchStatusTimeout(t *testing.T) {
	provider := NewSimulatedProvider(1)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := provider.FetchStatus(ctx, "ORD-9001")
	if err == nil {
		t.Fatal("esperava timeout, obteve nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("esperava context.DeadlineExceeded, obteve: %v", err)
	}
}

func TestSimulatedProvider_FetchStatusSuccess(t *testing.T) {
	provider := NewSimulatedProvider(2)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := provider.FetchStatus(ctx, "ORD-9002")
	if err != nil {
		t.Fatalf("não esperava erro, obteve: %v", err)
	}
	if result.ID != "ORD-9002" {
		t.Fatalf("esperava ID ORD-9002, obteve: %s", result.ID)
	}
	if result.Status == "" {
		t.Fatal("esperava status não vazio")
	}
	if result.Duration < 200*time.Millisecond {
		t.Fatalf("esperava delay mínimo de 200ms, obteve %s", result.Duration)
	}
}

func TestSimulatedProvider_FetchStatusProviderError(t *testing.T) {
	provider := NewSimulatedProvider(3)
	var lastErr error

	for i := 0; i < 20; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := provider.FetchStatus(ctx, "ORD-9003")
		cancel()
		lastErr = err
		if err == nil {
			continue
		}
		var providerErr *OrderProviderError
		if errors.As(err, &providerErr) {
			return
		}
		t.Fatalf("esperava erro do provedor, obteve: %T %v", err, err)
	}

	if lastErr == nil {
		t.Fatal("não obteve erro de provedor esperado dentro de 20 tentativas")
	}
}
