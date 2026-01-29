package handler_test

import (
	"TodoApp/handler"
	"TodoApp/model"
	"TodoApp/store"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIntegrationTest(t *testing.T) (*chi.Mux, func()) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err, "Failed to connect to test database")

	_, err = db.Exec(context.Background(), "TRUNCATE TABLE todos RESTART IDENTITY")
	require.NoError(t, err, "Failed to truncate todos table")

	postgresStore := store.NewTestPostgresStore(db)
	todoHandler := handler.NewTodoHandler(postgresStore)

	r := chi.NewRouter()
	r.Get("/todos", todoHandler.GetTodos)
	r.Post("/todos", todoHandler.CreateTodo)
	r.Get("/todos/{id}", todoHandler.GetTodo)
	r.Put("/todos/{id}", todoHandler.UpdateTodo)
	r.Delete("/todos/{id}", todoHandler.DeleteTodo)

	teardown := func() {
		db.Close()
	}

	return r, teardown
}

func TestTodoAPI_Integration(t *testing.T) {
	router, teardown := setupIntegrationTest(t)
	defer teardown()

	var createdTodo model.Todo

	t.Run("Create Todo", func(t *testing.T) {
		newTodo := `{"title":"Integration Test Todo", "completed":false}`
		req := httptest.NewRequest("POST", "/todos", bytes.NewBufferString(newTodo))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		err := json.NewDecoder(rr.Body).Decode(&createdTodo)
		require.NoError(t, err)
		assert.Equal(t, "Integration Test Todo", createdTodo.Title)
		assert.Equal(t, 1, createdTodo.ID)
	})

	t.Run("Get One Todo", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/todos/1", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var fetchedTodo model.Todo
		err := json.NewDecoder(rr.Body).Decode(&fetchedTodo)
		require.NoError(t, err)
		assert.Equal(t, createdTodo, fetchedTodo)
	})

	t.Run("Update Todo", func(t *testing.T) {
		updatedTodoJSON := `{"title":"Updated Integration Todo", "completed":true}`
		req := httptest.NewRequest("PUT", "/todos/1", bytes.NewBufferString(updatedTodoJSON))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var updatedTodo model.Todo
		err := json.NewDecoder(rr.Body).Decode(&updatedTodo)
		require.NoError(t, err)
		assert.Equal(t, "Updated Integration Todo", updatedTodo.Title)
		assert.True(t, updatedTodo.Completed)

		reqGet := httptest.NewRequest("GET", "/todos/1", nil)
		rrGet := httptest.NewRecorder()
		router.ServeHTTP(rrGet, reqGet)
		var verifiedTodo model.Todo
		json.NewDecoder(rrGet.Body).Decode(&verifiedTodo)
		assert.Equal(t, updatedTodo, verifiedTodo)
	})

	t.Run("Delete Todo", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/todos/1", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		reqGet := httptest.NewRequest("GET", "/todos/1", nil)
		rrGet := httptest.NewRecorder()
		router.ServeHTTP(rrGet, reqGet)
		assert.Equal(t, http.StatusNotFound, rrGet.Code)
	})
}
