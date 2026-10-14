package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type healthResponse struct {
	Status string `json:"status"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}

func main() {
	mux := http.NewServeMux()                    // ルーターを作る
	mux.HandleFunc("GET /health", healthHandler) // "GET /health" を healthHandler に結びつける

	log.Println("server started on :8080")
	http.ListenAndServe(":8080", mux) // ListenAndServe を呼び、エラーなら log.Fatal
}
