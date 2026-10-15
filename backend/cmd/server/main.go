package main

import (
	"context"
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
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./todo.db"
	}

	st, err := store.NewSQLite(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer st.Close()
	if err := st.Migrate(context.Background()); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	mux := http.NewServeMux()
	handler.NewTodoHandler(st).Register(mux) // handler は変更なし

	log.Printf("server started on :%s (db: %s)", port, dbPath)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
