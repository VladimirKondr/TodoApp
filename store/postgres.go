package store

import (
	"TodoApp/model"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TodoStore определяет методы для работы с хранилищем задач
type TodoStore interface {
	CreateTodo(ctx context.Context, todo *model.Todo) error
	GetTodos(ctx context.Context) ([]model.Todo, error)
	GetTodoByID(ctx context.Context, id int) (*model.Todo, error)
	UpdateTodo(ctx context.Context, todo *model.Todo) error
	DeleteTodo(ctx context.Context, id int) error
}

// PostgresStore реализует интерфейс TodoStore для PostgreSQL
type PostgresStore struct {
	db *pgxpool.Pool
}

// NewPostgresStore создает новый экземпляр PostgresStore
func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &PostgresStore{db: pool}, nil
}

// Close закрывает соединение с базой данных
func (s *PostgresStore) Close() {
	s.db.Close()
}

func (s *PostgresStore) CreateTodo(ctx context.Context, todo *model.Todo) error {
	query := "INSERT INTO todos (title, completed) VALUES ($1, $2) RETURNING id"
	return s.db.QueryRow(ctx, query, todo.Title, todo.Completed).Scan(&todo.ID)
}

func (s *PostgresStore) GetTodos(ctx context.Context) ([]model.Todo, error) {
	rows, err := s.db.Query(ctx, "SELECT id, title, completed FROM todos ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []model.Todo
	for rows.Next() {
		var t model.Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, nil
}

func (s *PostgresStore) GetTodoByID(ctx context.Context, id int) (*model.Todo, error) {
	var t model.Todo
	query := "SELECT id, title, completed FROM todos WHERE id = $1"
	err := s.db.QueryRow(ctx, query, id).Scan(&t.ID, &t.Title, &t.Completed)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *PostgresStore) UpdateTodo(ctx context.Context, todo *model.Todo) error {
	query := "UPDATE todos SET title = $1, completed = $2 WHERE id = $3"
	_, err := s.db.Exec(ctx, query, todo.Title, todo.Completed, todo.ID)
	return err
}

func (s *PostgresStore) DeleteTodo(ctx context.Context, id int) error {
	query := "DELETE FROM todos WHERE id = $1"
	_, err := s.db.Exec(ctx, query, id)
	return err
}

func NewTestPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: db}
}
