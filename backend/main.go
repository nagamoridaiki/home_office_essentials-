package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var (
	todos  = []Todo{}
	nextID = 1
	mu     sync.Mutex
)

// CORSミドルウェア
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	// Go 1.22+ ルーティング記法
	mux.HandleFunc("GET /api/todos", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todos)
	})

	mux.HandleFunc("POST /api/todos", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}
		mu.Lock()
		todo := Todo{ID: nextID, Title: req.Title, Completed: false}
		nextID++
		todos = append(todos, todo)
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(todo)
	})

	mux.HandleFunc("DELETE /api/todos/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		mu.Lock()
		newTodos := []Todo{}
		for _, t := range todos {
			if t.ID != id {
				newTodos = append(newTodos, t)
			}
		}
		todos = newTodos
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	fmt.Println("Backend running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}
