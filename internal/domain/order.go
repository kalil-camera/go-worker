package domain

import (
	"fmt"
	"time"
)

type OrderID string

type NodeID string

type OrderProcessingResult struct {
	ID       OrderID
	Status   string
	Duration time.Duration
	Err      error
}

type NodeProposal struct {
	NodeID    NodeID
	OrderID   OrderID
	Cost      float64
	Transit   time.Duration
	Available bool
	Err       error
}

type FulfillmentResult struct {
	OrderID       OrderID
	BestProposal  *NodeProposal
	NodeProposals []NodeProposal
	Errors        []error
}

type NodeTimeoutError struct {
	NodeID NodeID
}

func (e NodeTimeoutError) Error() string {
	return fmt.Sprintf("nó %s: tempo limite excedido na consulta de estoque", e.NodeID)
}

type InventoryNotAvailableError struct {
	OrderID OrderID
	NodeID  NodeID
	Reason  string
}

func (e InventoryNotAvailableError) Error() string {
	if e.NodeID == "" {
		return fmt.Sprintf("pedido %s: estoque não disponível (%s)", e.OrderID, e.Reason)
	}
	return fmt.Sprintf("pedido %s no nó %s: estoque não disponível (%s)", e.OrderID, e.NodeID, e.Reason)
}
