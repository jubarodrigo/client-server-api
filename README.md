# client-server-api

Desafio: cliente e servidor HTTP em Go com cotação do dólar, contextos com timeout e persistência em SQLite.

## Como executar

1. **Subir o servidor** (porta 8080, endpoint `/cotacao`):

   ```bash
   go run -tags server .
   ```
   ou, após compilar: `./server`

2. **Em outro terminal, rodar o cliente**:

   ```bash
   go run -tags client .
   ```
   ou: `./client`

O cliente grava a cotação em `cotacao.txt` no formato `Dólar: {valor}`.

## Requisitos atendidos

- **server.go**: consome a API USD-BRL (AwesomeAPI), retorna JSON com o `bid`; timeouts: 200ms para a API e 10ms para persistir no SQLite; erros de contexto registrados em log.
- **client.go**: requisição ao servidor com timeout de 300ms; recebe apenas o valor do câmbio (bid); salva em `cotacao.txt`; erros de contexto registrados em log.
- **Contextos**: os 3 contextos (API 200ms, DB 10ms, cliente 300ms) retornam erro nos logs quando o tempo é insuficiente.