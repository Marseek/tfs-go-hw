package main

import (
	"log"
	"net/http"

	"github.com/Marseek/tfs-go-hw/home_work_04/domain"
	"github.com/Marseek/tfs-go-hw/home_work_04/handlers"
	"github.com/Marseek/tfs-go-hw/home_work_04/repository"
	"github.com/go-chi/chi/v5"
)

func main() {
	storage := domain.MessageStorage{UserNamesAndPass: map[string]string{}, ToAllMessages: []domain.Messages{}, PersonalMessages: map[string][]domain.Messages{}}
	rep := repository.NewRepository(&storage)

	handler := handlers.NewChatHandlers(rep)

	r := chi.NewRouter()
	r.Mount("/", handler.Routes())
	log.Fatal(http.ListenAndServe(":5000", r))
}
