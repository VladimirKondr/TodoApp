package handler

import (
	"TodoApp/model"
	"TodoApp/store"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type TodoHandler struct {
	store store.TodoStore
}

func NewTodoHandler(s store.TodoStore) *TodoHandler {
	return &TodoHandler{store: s}
}

func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := h.store.GetTodos(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, todos)
}

func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todo model.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.store.CreateTodo(r.Context(), &todo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusCreated, todo)
}

func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	todo, err := h.store.GetTodoByID(r.Context(), id)
	if err != nil {
		// Здесь можно было бы проверить на pgx.ErrNoRows для 404
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	respondJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var todo model.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	todo.ID = id

	if err := h.store.UpdateTodo(r.Context(), &todo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteTodo(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}
