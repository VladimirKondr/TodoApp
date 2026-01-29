package handler

import (
	"TodoApp/model"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTodoStore - это мок-реализация интерфейса TodoStore для тестов
type MockTodoStore struct {
	todos   []model.Todo
	err     error
	counter int // Для имитации автоинкремента ID
}

func (m *MockTodoStore) CreateTodo(ctx context.Context, todo *model.Todo) error {
	if m.err != nil {
		return m.err
	}
	m.counter++
	todo.ID = m.counter
	m.todos = append(m.todos, *todo)
	return nil
}

func (m *MockTodoStore) GetTodos(ctx context.Context) ([]model.Todo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.todos, nil
}

func (m *MockTodoStore) GetTodoByID(ctx context.Context, id int) (*model.Todo, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, t := range m.todos {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockTodoStore) UpdateTodo(ctx context.Context, todo *model.Todo) error {
	if m.err != nil {
		return m.err
	}
	for i, t := range m.todos {
		if t.ID == todo.ID {
			m.todos[i] = *todo
			return nil
		}
	}
	return errors.New("not found on update")
}

func (m *MockTodoStore) DeleteTodo(ctx context.Context, id int) error {
	if m.err != nil {
		return m.err
	}
	for i, t := range m.todos {
		if t.ID == id {
			m.todos = append(m.todos[:i], m.todos[i+1:]...)
			return nil
		}
	}
	return errors.New("not found on delete")
}

// --- СУЩЕСТВУЮЩИЕ ТЕСТЫ ---

func TestGetTodos(t *testing.T) {
	// Настройка
	mockStore := &MockTodoStore{
		todos: []model.Todo{
			{ID: 1, Title: "Test Todo 1", Completed: false},
		},
		counter: 1,
	}
	handler := NewTodoHandler(mockStore)
	req := httptest.NewRequest("GET", "/todos", nil)
	rr := httptest.NewRecorder()

	// Выполнение
	handler.GetTodos(rr, req)

	// Проверка
	assert.Equal(t, http.StatusOK, rr.Code)

	var todos []model.Todo
	err := json.NewDecoder(rr.Body).Decode(&todos)
	require.NoError(t, err)
	assert.Len(t, todos, 1)
	assert.Equal(t, "Test Todo 1", todos[0].Title)
}

func TestCreateTodo(t *testing.T) {
	// Настройка
	mockStore := &MockTodoStore{}
	handler := NewTodoHandler(mockStore)

	todo := model.Todo{Title: "New Todo", Completed: false}
	body, _ := json.Marshal(todo)
	req := httptest.NewRequest("POST", "/todos", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	// Выполнение
	handler.CreateTodo(rr, req)

	// Проверка
	assert.Equal(t, http.StatusCreated, rr.Code)

	var createdTodo model.Todo
	err := json.NewDecoder(rr.Body).Decode(&createdTodo)
	require.NoError(t, err)
	assert.Equal(t, "New Todo", createdTodo.Title)
	assert.Equal(t, 1, createdTodo.ID) // ID должен быть присвоен моком
}

func TestGetTodoByID_NotFound(t *testing.T) {
	// Настройка
	mockStore := &MockTodoStore{} // Пустое хранилище
	handler := NewTodoHandler(mockStore)

	req := httptest.NewRequest("GET", "/todos/99", nil)
	rr := httptest.NewRecorder()

	// Для извлечения URL-параметра нам нужен роутер chi
	router := chi.NewRouter()
	router.Get("/todos/{id}", handler.GetTodo)

	// Выполнение
	router.ServeHTTP(rr, req)

	// Проверка
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- НОВЫЕ ТЕСТЫ ---

func TestUpdateTodo(t *testing.T) {
	// Настройка
	mockStore := &MockTodoStore{
		todos: []model.Todo{
			{ID: 1, Title: "Old Title", Completed: false},
		},
		counter: 1,
	}
	handler := NewTodoHandler(mockStore)

	updatedTodo := model.Todo{Title: "New Title", Completed: true}
	body, _ := json.Marshal(updatedTodo)
	req := httptest.NewRequest("PUT", "/todos/1", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Put("/todos/{id}", handler.UpdateTodo)

	// Выполнение
	router.ServeHTTP(rr, req)

	// Проверка
	assert.Equal(t, http.StatusOK, rr.Code)

	var responseTodo model.Todo
	err := json.NewDecoder(rr.Body).Decode(&responseTodo)
	require.NoError(t, err)
	assert.Equal(t, "New Title", responseTodo.Title)
	assert.True(t, responseTodo.Completed)
	assert.Equal(t, 1, mockStore.todos[0].ID) // Убедимся, что в моке тоже обновилось
	assert.Equal(t, "New Title", mockStore.todos[0].Title)
}

func TestDeleteTodo(t *testing.T) {
	// Настройка
	mockStore := &MockTodoStore{
		todos: []model.Todo{
			{ID: 1, Title: "To be deleted", Completed: false},
		},
		counter: 1,
	}
	handler := NewTodoHandler(mockStore)

	req := httptest.NewRequest("DELETE", "/todos/1", nil)
	rr := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Delete("/todos/{id}", handler.DeleteTodo)

	// Выполнение
	router.ServeHTTP(rr, req)

	// Проверка
	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Len(t, mockStore.todos, 0) // Задача должна быть удалена из мока
}
