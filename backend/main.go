package main

import (
	"fmt"
	"log"
	"net/http"

	"backend/controller"
	"backend/repository"
	"backend/usecase"
)

func main() {
	todoRepo := repository.NewMemoryTodoRepository()
	todoUsecase := usecase.NewTodoUsecase(todoRepo)
	todoController := controller.NewTodoController(todoUsecase)

	mux := http.NewServeMux()
	todoController.Register(mux)

	fmt.Println("Backend running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", controller.EnableCORS(mux)))
}
