package repository

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go-worker/internal/domain"
)

type NodeSimulator struct {
	rng *rand.Rand
}

type NodeInfo struct {
	ID        domain.NodeID
	Available bool
}

type SimulatedProvider struct {
	rng *rand.Rand
}

func NewNodeSimulator(seed int64) *NodeSimulator {
	return &NodeSimulator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

func NewSimulatedProvider(seed int64) *SimulatedProvider {
	return &SimulatedProvider{
		rng: rand.New(rand.NewSource(seed)),
	}
}

func (p *NodeSimulator) QueryNode(ctx context.Context, nodeID domain.NodeID, orderID domain.OrderID) (domain.NodeProposal, error) {
	delay := p.randomDelay(nodeID)
	cost := 5.0 + float64(p.rng.Intn(26))
	availability := p.rng.Float32() >= 0.15

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return domain.NodeProposal{NodeID: nodeID, OrderID: orderID}, ctx.Err()
	}

	if !availability {
		return domain.NodeProposal{NodeID: nodeID, OrderID: orderID, Cost: cost, Transit: delay, Available: false}, domain.InventoryNotAvailableError{
			OrderID: orderID,
			NodeID:  nodeID,
			Reason:  "estoque inexistente no nó",
		}
	}

	return domain.NodeProposal{
		NodeID:    nodeID,
		OrderID:   orderID,
		Cost:      cost,
		Transit:   delay,
		Available: true,
	}, nil
}

func (p *NodeSimulator) randomDelay(nodeID domain.NodeID) time.Duration {
	id := string(nodeID)
	if strings.HasPrefix(id, "STORE-01") {
		return time.Duration(100+p.rng.Intn(180)) * time.Millisecond
	}
	if strings.HasPrefix(id, "STORE-02") {
		return time.Duration(120+p.rng.Intn(160)) * time.Millisecond
	}
	if strings.HasPrefix(id, "STORE-03") {
		return time.Duration(90+p.rng.Intn(200)) * time.Millisecond
	}
	if strings.HasPrefix(id, "CD-") {
		return time.Duration(220+p.rng.Intn(200)) * time.Millisecond
	}
	return time.Duration(100+p.rng.Intn(260)) * time.Millisecond
}

func (p *NodeSimulator) DescribeNode(nodeID domain.NodeID) string {
	return fmt.Sprintf("Nó %s", nodeID)
}

func (p *SimulatedProvider) FetchStatus(ctx context.Context, orderID domain.OrderID) (domain.OrderProcessingResult, error) {
	delay := time.Duration(200+p.rng.Intn(1300)) * time.Millisecond
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

func (p *SimulatedProvider) DescribeProvider() string {
	return "SimulatedProvider"
}

func (p *SimulatedProvider) QueryNode(ctx context.Context, nodeID domain.NodeID, orderID domain.OrderID) (domain.NodeProposal, error) {
	proposal, err := p.FetchStatus(ctx, orderID)
	return domain.NodeProposal{
		NodeID:    nodeID,
		OrderID:   orderID,
		Cost:      0,
		Transit:   proposal.Duration,
		Available: err == nil,
	}, err
}

type OrderProviderError struct {
	OrderID domain.OrderID
	Reason  string
}

func (e OrderProviderError) Error() string {
	return "provedor externo: pedido " + string(e.OrderID) + ": " + e.Reason
}
