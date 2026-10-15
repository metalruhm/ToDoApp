package main

import (
	"log"
	"net/http"
	"os"

	"todoapp/backend/internal/handler"
	"todoapp/backend/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	st := store.NewMemory() // Section 6 で SQLite に差し替える
	h := handler.NewTodoHandler(st)

	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("server started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
