package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type InitializePaymentResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

func main() {
	port := ":8081"
	http.HandleFunc("/transaction/initialize", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Simulate slight processing delay
		time.Sleep(1 * time.Millisecond)

		resp := InitializePaymentResponse{
			Status:  true,
			Message: "Authorization URL created",
		}
		resp.Data.AuthorizationURL = "https://checkout.paystack.com/mock_auth_url"
		resp.Data.AccessCode = "mock_access_code"
		resp.Data.Reference = fmt.Sprintf("mock_ref_%d", time.Now().UnixNano())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)

		log.Printf("Processed mock payment initialization: %s", resp.Data.Reference)
	})

	log.Printf("Mock Paystack server starting on %s...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
