package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"backend/domain"
	"backend/usecase"
)

type TodoController struct {
	usecase *usecase.TodoUsecase
}

func NewTodoController(u *usecase.TodoUsecase) *TodoController {
	return &TodoController{usecase: u}
}

func (c *TodoController) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/todos", c.list)
	mux.HandleFunc("POST /api/todos", c.create)
	mux.HandleFunc("DELETE /api/todos/{id}", c.delete)
}

type todoResponse struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func toResponse(t domain.Todo) todoResponse {
	return todoResponse{ID: t.ID, Title: t.Title, Completed: t.Completed}
}

func (c *TodoController) list(w http.ResponseWriter, r *http.Request) {
	todos, err := c.usecase.List()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// nilだとJSONが null になるため、空でも [] を返す
	res := make([]todoResponse, 0, len(todos))
	for _, t := range todos {
		res = append(res, toResponse(t))
	}
	writeJSON(w, http.StatusOK, res)
}

func (c *TodoController) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	todo, err := c.usecase.Create(req.Title)
	if errors.Is(err, domain.ErrEmptyTitle) {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, toResponse(todo))
}

func (c *TodoController) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	if err := c.usecase.Delete(id); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
