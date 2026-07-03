package logistics

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-worker/internal/domain"
	"go-worker/internal/repository"
)

func TestFindBestFulfillment_Timeout(t *testing.T) {
	simulator := repository.NewNodeSimulator(1)
	planner := NewFulfillmentPlanner(simulator, 50*time.Millisecond)

	result, err := planner.FindBestFulfillment(context.Background(), "ORD-7001", []domain.NodeID{"CD-01", "STORE-01"})
	if err == nil {
		t.Fatal("esperava erro de timeout ou falta de proposta, obteve nil")
	}
	if result.BestProposal != nil {
		t.Fatal("esperava nenhuma proposta vencedora devido ao timeout")
	}
	var timeoutErr domain.NodeTimeoutError
	found := false
	for _, e := range result.Errors {
		if errors.As(e, &timeoutErr) {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("esperava pelo menos um NodeTimeoutError nas propostas")
	}
}
