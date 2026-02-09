//go:build client

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	clientTimeout = 300 * time.Millisecond
	serverURL     = "http://localhost:8080/cotacao"
)

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {
	// Contexto com timeout máximo de 300ms para receber o resultado do server
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		log.Fatalf("[ERRO] contexto: falha ao criar requisição: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("[ERRO] contexto: timeout de %v para receber resultado do server.go", clientTimeout)
		} else {
			log.Printf("[ERRO] contexto: falha na requisição: %v", err)
		}
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var cotacao CotacaoResponse
	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		log.Fatalf("[ERRO] contexto: falha ao decodificar resposta: %v", err)
	}

	if cotacao.Bid == "" {
		log.Fatal("[ERRO] contexto: valor de cotação (bid) vazio")
	}

	// Salvar em cotacao.txt no formato: Dólar: {valor}
	conteudo := "Dólar: " + cotacao.Bid + "\n"
	if err := os.WriteFile("cotacao.txt", []byte(conteudo), 0644); err != nil {
		log.Fatalf("[ERRO] contexto: falha ao salvar arquivo: %v", err)
	}

	log.Printf("Cotação salva: Dólar: %s", cotacao.Bid)
}
