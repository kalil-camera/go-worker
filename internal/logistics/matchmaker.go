package logistics

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"go-worker/internal/domain"
)

type NodeRepository interface {
	QueryNode(ctx context.Context, nodeID domain.NodeID, orderID domain.OrderID) (domain.NodeProposal, error)
}

type FulfillmentPlanner struct {
	repository NodeRepository
	timeout    time.Duration
}

func NewFulfillmentPlanner(repo NodeRepository, timeout time.Duration) *FulfillmentPlanner {
	return &FulfillmentPlanner{
		repository: repo,
		timeout:    timeout,
	}
}

func (p *FulfillmentPlanner) FindBestFulfillment(ctx context.Context, orderID domain.OrderID, nodes []domain.NodeID) (domain.FulfillmentResult, error) {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	responseCh := make(chan domain.NodeProposal, len(nodes))

	var wg sync.WaitGroup
	for _, nodeID := range nodes {
		wg.Add(1)
		go func(nodeID domain.NodeID) {
			defer wg.Done()

			nodeCtx, nodeCancel := context.WithTimeout(ctx, p.timeout)
			defer nodeCancel()

			proposal, err := p.repository.QueryNode(nodeCtx, nodeID, orderID)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					err = domain.NodeTimeoutError{NodeID: nodeID}
				}
				responseCh <- domain.NodeProposal{
					NodeID:    nodeID,
					OrderID:   orderID,
					Available: false,
					Err:       err,
				}
				return
			}
			proposal.NodeID = nodeID
			proposal.OrderID = orderID
			responseCh <- proposal
		}(nodeID)
	}

	go func() {
		wg.Wait()
		close(responseCh)
	}()

	var proposals []domain.NodeProposal
	var validProposals []domain.NodeProposal
	var collectedErrors []error
	for proposal := range responseCh {
		proposals = append(proposals, proposal)
		if proposal.Err != nil {
			collectedErrors = append(collectedErrors, proposal.Err)
			continue
		}
		validProposals = append(validProposals, proposal)
	}

	if len(validProposals) == 0 {
		return domain.FulfillmentResult{OrderID: orderID, NodeProposals: proposals, Errors: collectedErrors}, fmt.Errorf("nenhuma proposta válida recebida")
	}

	sort.Slice(validProposals, func(i, j int) bool {
		scoreI := validProposals[i].Cost + float64(validProposals[i].Transit.Milliseconds())/100.0
		scoreJ := validProposals[j].Cost + float64(validProposals[j].Transit.Milliseconds())/100.0
		return scoreI < scoreJ
	})

	best := validProposals[0]
	return domain.FulfillmentResult{OrderID: orderID, BestProposal: &best, NodeProposals: proposals, Errors: collectedErrors}, nil
}
