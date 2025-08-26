package main

import (
	"fmt"
	"net/http"

	"github.com/PavelBradnitski/WbTechL3.1/internal/handler"
	"github.com/PavelBradnitski/WbTechL3.1/internal/repository"
	"github.com/PavelBradnitski/WbTechL3.1/internal/service"
)

func main() {
	repo := repository.NewMemoryRepo()
	svc := service.NewNotificationService(repo)
	h := handler.NewNotificationHandler(svc)

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", h.Routes())
}
