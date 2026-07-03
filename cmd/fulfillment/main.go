package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-worker/internal/domain"
	"go-worker/internal/logistics"
	"go-worker/internal/repository"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	simulator := repository.NewNodeSimulator(time.Now().UnixNano())
	planner := logistics.NewFulfillmentPlanner(simulator, 400*time.Millisecond)

	orders := []domain.OrderID{
		domain.OrderID("ORD-5001"),
		domain.OrderID("ORD-5002"),
		domain.OrderID("ORD-5003"),
		domain.OrderID("ORD-5004"),
		domain.OrderID("ORD-5005"),
		domain.OrderID("ORD-5006"),
	}
	storeNodes := []domain.NodeID{"STORE-01", "STORE-02", "STORE-03"}
	fallbackNode := []domain.NodeID{"CD-01"}

	totalOrders := len(orders)
	successfulPrimary := 0
	fallbackUsed := 0
	fallbackSucceeded := 0
	failedOrders := 0

	for _, order := range orders {
		fmt.Printf("\n--- Iniciando leilão de fulfillment para pedido %s ---\n", order)

		result, err := planner.FindBestFulfillment(ctx, order, storeNodes)
		if err != nil {
			log.Printf("leilão primário de lojas falhou: %v", err)
		}

		usedFallback := false
		if result.BestProposal == nil {
			usedFallback = true
			fallbackUsed++
			fmt.Println("Nenhuma loja respondeu a tempo. Executando fallback para o centro de distribuição.")
			result, err = planner.FindBestFulfillment(ctx, order, fallbackNode)
			if err != nil {
				log.Printf("falha no fallback para CD: %v", err)
			}
		}

		if result.BestProposal != nil {
			if usedFallback {
				fallbackSucceeded++
			} else {
				successfulPrimary++
			}
		} else {
			failedOrders++
		}

		fmt.Println("Resultado do leilão:")
		fmt.Println("Propostas por nó:")
		fmt.Printf("  %-10s %-8s %-10s %-12s %s\n", "NÓ", "CUSTO", "PRAZO", "DISPONÍVEL", "STATUS")
		fmt.Println("  --------------------------------------------------------------")
		for _, proposal := range result.NodeProposals {
			available := "não"
			status := "OK"
			if proposal.Available {
				available = "sim"
			}
			if proposal.Err != nil {
				switch {
				case errors.As(proposal.Err, &domain.NodeTimeoutError{}):
					status = "TIMEOUT"
				case errors.As(proposal.Err, &domain.InventoryNotAvailableError{}):
					status = "NO_STOCK"
				default:
					status = "ERROR"
				}
			}
			fmt.Printf("  %-10s R$ %-6.2f %-10s %-12s %s\n", proposal.NodeID, proposal.Cost, proposal.Transit, available, status)
		}
		fmt.Println("\nMelhor proposta:")
		if result.BestProposal != nil {
			fmt.Printf("  NÓ: %s\n", result.BestProposal.NodeID)
			fmt.Printf("  CUSTO: R$ %.2f\n", result.BestProposal.Cost)
			fmt.Printf("  PRAZO: %s\n", result.BestProposal.Transit)
		} else {
			fmt.Println("  Nenhuma proposta válida encontrada.")
		}
		if len(result.Errors) > 0 {
			fmt.Println("Erros registrados:")
			for _, e := range result.Errors {
				fmt.Printf("  - %v\n", e)
			}
		}
	}

	fmt.Println("\n=== Resumo final ===")
	fmt.Printf("Pedidos processados: %d\n", totalOrders)
	fmt.Printf("Fulfillments bem-sucedidos em lojas físicas: %d\n", successfulPrimary)
	fmt.Printf("Fallbacks para CD usados: %d\n", fallbackUsed)
	fmt.Printf("Fallbacks bem-sucedidos: %d\n", fallbackSucceeded)
	fmt.Printf("Pedidos sem proposta válida: %d\n", failedOrders)

	if err := ctx.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "shutdown gracioso capturado:", err)
	}
}
