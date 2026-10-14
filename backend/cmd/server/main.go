package main

import (
	"fmt"
	"log"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// ここで "ok" を書き込む
	fmt.Fprint(w, "ok")
}

func main() {
	mux := http.NewServeMux()                    // ルーターを作る
	mux.HandleFunc("GET /health", healthHandler) // "GET /health" を healthHandler に結びつける

	log.Println("server started on :8080")
	http.ListenAndServe(":8080", mux) // ListenAndServe を呼び、エラーなら log.Fatal
}
