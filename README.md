# Omnichannel Fulfillment Matchmaker CLI

Uma CLI em Go que simula um roteador omnicanal de leilão de estoque para decisão de despacho em um varejo.

## Como executar

1. Abra o terminal em `c:\Users\Admin\Desktop\projetos\go-worker`
2. Execute:

```powershell
 go run ./cmd/fulfillment
```

Para rodar testes:

```powershell
 go test ./...
```

## Estrutura de pastas

- `cmd/fulfillment`: ponto de entrada do roteador omnicanal.
- `internal/domain`: tipos de domínio e erros customizados do fulfillment.
- `internal/logistics`: lógica de leilão de estoque, concorrência e seleção do melhor nó.
- `internal/repository`: simulação dos nós de atendimento (CD e lojas físicas).

## Justificativa arquitetural

A aplicação segue uma arquitetura limpa simplificada para separar domínio, regras de negócio e infraestrutura.

- `internal/domain` contém modelos e erros de negócio.
- `internal/logistics` realiza o fan-out das consultas aos nós, coleta propostas com channels seguros e decide o vencedor.
- `internal/repository` simula a infraestrutura de estoque com latência e disponibilidade variáveis.

## O modelo de negócio

Quando um pedido é recebido, o sistema dispara consultas simultâneas para múltiplos nós de atendimento:

- 1 centro de distribuição principal
- 3 lojas físicas próximas

Cada nó envia uma proposta com custo de frete e tempo de atendimento. O sistema escolhe a melhor combinação de preço e tempo dentro de um timeout estrito de `400ms`.

## Concorrência, contexto e shutdown

- O padrão `fan-out/fan-in` é usado com `goroutines` para consultar todos os nós ao mesmo tempo.
- Cada consulta recebe um `context.WithTimeout` para garantir que um nó lento seja cancelado e ignorado.
- As propostas chegam em um channel e são processadas de forma segura, evitando race conditions.
- Um `signal.NotifyContext` é usado no comando principal para capturar `SIGTERM` e fazer shutdown gracioso.

## Como o leilão funciona

- Se um nó não responder dentro de `400ms`, sua proposta é descartada e um erro de timeout é registrado.
- O melhor nó é escolhido pela menor combinação de `cost + transit time`.
- Se um nó não tiver estoque disponível, o erro `ErrInventoryNotAvailable` é retornado.
