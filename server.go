//go:build server

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

const (
	apiTimeout = 200 * time.Millisecond
	dbTimeout  = 10 * time.Millisecond
)

type USDBRL struct {
	Bid string `json:"bid"`
}

type APIResponse struct {
	USDBRL USDBRL `json:"USDBRL"`
}

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {
	db, err := sql.Open("sqlite", "cotacoes.db")
	if err != nil {
		log.Fatal("erro ao abrir banco:", err)
	}
	defer db.Close()

	if err := criarTabela(db); err != nil {
		log.Fatal("erro ao criar tabela:", err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		handlerCotacao(w, r, db)
	})

	log.Println("Servidor ouvindo na porta 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func criarTabela(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cotacoes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bid TEXT NOT NULL,
			criado_em DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func handlerCotacao(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// 1. Buscar cotação na API externa (timeout 200ms)
	ctxAPI, cancelAPI := context.WithTimeout(r.Context(), apiTimeout)
	defer cancelAPI()

	req, err := http.NewRequestWithContext(ctxAPI, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Printf("[ERRO] contexto: falha ao criar requisição à API: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctxAPI.Err() == context.DeadlineExceeded {
			log.Printf("[ERRO] contexto: timeout de %v ao chamar API de cotação do dólar", apiTimeout)
		} else {
			log.Printf("[ERRO] contexto: falha ao chamar API: %v", err)
		}
		http.Error(w, err.Error(), http.StatusGatewayTimeout)
		return
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		log.Printf("[ERRO] contexto: falha ao decodificar resposta da API: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bid := apiResp.USDBRL.Bid
	if bid == "" {
		log.Printf("[ERRO] contexto: campo bid vazio na resposta da API")
		http.Error(w, "bid vazio", http.StatusInternalServerError)
		return
	}

	// 2. Persistir no SQLite (timeout 10ms)
	ctxDB, cancelDB := context.WithTimeout(r.Context(), dbTimeout)
	defer cancelDB()

	_, err = db.ExecContext(ctxDB, "INSERT INTO cotacoes (bid) VALUES (?)", bid)
	if err != nil {
		if ctxDB.Err() == context.DeadlineExceeded {
			log.Printf("[ERRO] contexto: timeout de %v ao persistir dados no banco", dbTimeout)
		} else {
			log.Printf("[ERRO] contexto: falha ao persistir no banco: %v", err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Retornar apenas o bid para o cliente
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CotacaoResponse{Bid: bid})
}
