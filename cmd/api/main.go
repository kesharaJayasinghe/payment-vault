package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/lib/pq" // Postgres Driver
	"github.com/kesharaJayasinghe/payment-vault/internal/payment"
)

type PaymentRequest struct {
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type PaymentResponse struct {
	Success       bool   `json:"success"`
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id,omitempty"`
	Error         string `json:"error,omitempty"`
}

func main() {
	// Connect to DB
	connStr := "postgres://vault:securepassword@localhost:5435/vault?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	provider := payment.NewMockProvider()

	http.HandleFunc("POST /charge", func(w http.ResponseWriter, r *http.Request) {
		// Extract idempotency key
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "Missing Idempotency-Key header", http.StatusBadRequest)
			return
		}

		// Check DB for existing transaction with this key
		var existingBody []byte
		err := db.QueryRow("SELECT response_body FROM payment_requests WHERE idempotency_key = $1", key).Scan(&existingBody)

		if err == nil && existingBody != nil {
			// Have processed this already. Return the SAVED response.
			log.Printf("Idempotency Hit: Returning saved response for %s", key)
			w.Header().Set("Content-Type", "application/json")
			w.Write(existingBody)
			return
		}

		// New request: Parse request body
		var req PaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Insert "STARTED" state
		// Insert *before* calling the provider to "claim" the key.
		_, err = db.Exec(`
			INSERT INTO payment_requests (idempotency_key, user_id, amount, currency, status)
			VALUES ($1, $2, $3, $4, 'STARTED')
		`, key, req.UserID, req.Amount, req.Currency)
		
		if err != nil {
			log.Printf("DB Insert Error: %v", err)
			// If insert fails (Duplicate Key Race Condition), return 409 Conflict
			http.Error(w, "Concurrent processing detected", http.StatusConflict)
			return
		}

		// call the payment provider
		txnID, chargeErr := provider.Charge(req.Amount, req.Currency)

		// Determine results
		resp := PaymentResponse{
			Success: chargeErr == nil,
			Status:  "SUCCEEDED",
		}
		if chargeErr != nil {
			resp.Status = "FAILED"
			resp.Error = chargeErr.Error()
		} else {
			resp.TransactionID = txnID
		}

		// Save response and update status in DB
		respBody, _ := json.Marshal(resp)
		_, err = db.Exec(`
			UPDATE payment_requests 
			SET status = $1, response_body = $2, updated_at = NOW() 
			WHERE idempotency_key = $3
		`, resp.Status, respBody, key)

		if err != nil {
			log.Printf("CRITICAL: Failed to save response for key %s: %v", key, err)
			// In a production env, raise a serious alert.
		}

		// Reply to user
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBody)
	})

	log.Println("Vault Payment Service listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}