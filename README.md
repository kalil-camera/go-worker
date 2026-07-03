# Go Worker CLI

Uma pequena CLI em Go que simula um verificador concorrente de status de pedidos para um varejista.

## Como executar

1. Abra o terminal em `c:\Users\Admin\Desktop\projetos\go-worker`
2. Execute:

```powershell
 go run ./cmd/order-status
```

Para rodar testes:

```powershell
 go test ./...
```

## Estrutura de pastas

- `cmd/order-status`: ponto de entrada da aplicação.
- `internal/domain`: tipos de domínio e estruturas de dados principais.
- `internal/service`: regra de negócio de checagem concorrente de pedidos.
- `internal/repository`: simulação do provedor externo de status.

## Justificativa arquitetural

A aplicação foi organizada em packages para separar a lógica de domínio, serviço e infraestrutura.

- `internal/domain` contém os modelos de dados usados pelo serviço.
- `internal/service` coordena a execução concorrente, agrupa resultados e aplica timeouts via `context`.
- `internal/repository` encapsula a simulação de um serviço externo com delay e taxa de falha.

## Concorrência e contexto

- As consultas são processadas em paralelo com `goroutines`.
- `sync.WaitGroup` garante que o fluxo principal aguarde todas as goroutines.
- Cada requisição utiliza `context.WithTimeout` para respeitar um timeout isolado.
- O código demonstra tratamento explícito de erros e customização de erro para timeouts e falhas do provedor.
