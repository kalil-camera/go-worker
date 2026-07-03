package repository

import (
    "context"
    "math/rand"
    "time"

    "go-worker/internal/domain"
)

type SimulatedProvider struct {
    rng *rand.Rand
}

func NewSimulatedProvider(seed int64) *SimulatedProvider {
    return &SimulatedProvider{
        rng: rand.New(rand.NewSource(seed)),
    }
}

func (p *SimulatedProvider) FetchStatus(ctx context.Context, orderID domain.OrderID) (domain.OrderProcessingResult, error) {
    delay := time.Duration(200+ p.rng.Intn(1300)) * time.Millisecond
    select {
    case <-time.After(delay):
    case <-ctx.Done():
        return domain.OrderProcessingResult{ID: orderID}, ctx.Err()
    }

    if p.rng.Float32() < 0.20 {
        return domain.OrderProcessingResult{ID: orderID, Duration: delay}, &OrderProviderError{
            OrderID: orderID,
            Reason:  "falha simulada no provedor externo",
        }
    }

    status := []string{"Processando", "Entregue", "Cancelado"}[p.rng.Intn(3)]
    return domain.OrderProcessingResult{
        ID:       orderID,
        Status:   status,
        Duration: delay,
    }, nil
}

type OrderProviderError struct {
    OrderID domain.OrderID
    Reason  string
}

func (e OrderProviderError) Error() string {
    return "proveedor externo: pedido " + string(e.OrderID) + ": " + e.Reason
}
