package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Marseek/tfs-go-hw/course/handlers"
	pkgpostgres "github.com/Marseek/tfs-go-hw/course/pkg/postgres"
	"github.com/Marseek/tfs-go-hw/course/repository"
	"github.com/Marseek/tfs-go-hw/course/service"
	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

func main() {
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	dsn := "postgres://jlexie:passwd@localhost:5442/fintech" +
		"?sslmode=disable"
	pool, err := pkgpostgres.NewPool(dsn, logger)
	if err != nil {
		logger.Fatal(err)
	}
	defer pool.Close()

	rep := repository.NewRepository(pool, logger)
	serv := service.NewRobotService(rep, logger)
	handler := handlers.NewParamsSetter(logger, serv)
	// query := `TRUNCATE TABLE orders`
	// pool.Exec(context.Background(), query)

	mainSrv := http.Server{}
	r := chi.NewRouter()
	r.Mount("/", handler.Routes())
	go func() {
		logger.Fatal(http.ListenAndServe(":5000", r))
	}()

	// Gracefull shutdown
	sigquit := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	signal.Notify(sigquit, syscall.SIGINT, syscall.SIGTERM)
	stopAppCh := make(chan struct{})
	go func() {
		log.Println("Captured signal: ", <-sigquit)
		log.Println("Gracefully shutting down server...")
		if err := mainSrv.Shutdown(context.Background()); err != nil {
			log.Println("Can't shutdown main server: ", err.Error())
		}
		stopAppCh <- struct{}{}
	}()
	<-stopAppCh
}
