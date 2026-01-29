package main

import (
	"TodoApp/handler"
	"TodoApp/store"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	dbStore, err := store.NewPostgresStore(databaseURL)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer dbStore.Close()

	todoHandler := handler.NewTodoHandler(dbStore)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/todos", todoHandler.GetTodos)
	r.Post("/todos", todoHandler.CreateTodo)
	r.Get("/todos/{id}", todoHandler.GetTodo)
	r.Put("/todos/{id}", todoHandler.UpdateTodo)
	r.Delete("/todos/{id}", todoHandler.DeleteTodo)

	log.Println("Starting server on :8081")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
