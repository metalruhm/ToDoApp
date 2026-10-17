package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todoapp/backend/internal/handler"
	"todoapp/backend/internal/store"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	port := getenv("PORT", "8080")
	dbPath := getenv("DB_PATH", "./todo.db")
	allowedOrigin := getenv("ALLOWED_ORIGIN", "http://localhost:3000")
	secureCookie := os.Getenv("COOKIE_SECURE") == "true"

	st, err := store.NewSQLite(dbPath)
	if err != nil {
		return err
	}
	defer st.Close()
	if err := st.Migrate(context.Background()); err != nil {
		return err
	}

	mux := http.NewServeMux()
	authH := handler.NewAuthHandler(st, secureCookie)
	authH.Register(mux)
	handler.NewTodoHandler(st).Register(mux, authH.RequireAuth)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler.CORS(allowedOrigin, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Ctrl+C / SIGTERM で、処理中のリクエストを待ってから終了する
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("server started on :%s (db: %s, origin: %s)", port, dbPath, allowedOrigin)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Println("server stopped")
	return nil
}
