package main

import (
	"awesomeeng/otel-demo/utils"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const serviceName = "user-service"

var (
	// db      datastore.DB
	srv     *http.Server
	userUrl string
	tracer  trace.Tracer
)

func setupServer() {
	router := mux.NewRouter()
	// router.HandleFunc("/users", createUser).Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/users/{userID}", getUser).Methods(http.MethodGet, http.MethodOptions)
	// router.HandleFunc("/users/{userID}", updateUser).Methods(http.MethodPut, http.MethodOptions)
	// router.Use(utils.LoggingMW)
	router.Use(otelmux.Middleware(serviceName))
	router.Use(utils.MetricMiddleware)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost},
	})

	srv = &http.Server{
		Addr:    ":8080",
		Handler: c.Handler(router),
	}

	log.Printf("User service running at: :8080")
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to setup http server: %v", err)
	}
}

// func initDB() {
// 	var err error
// 	if db, err = datastore.New(); err != nil {
// 		log.Fatalf("failed to initialize db: %v", err)
// 	}
// }

func main() {
	utils.InitTraceProvider()
	userUrl = os.Getenv("USER_URL")
	utils.InitMetricProvider(utils.InitServerMeter)
	// setup tracer
	// tp := config.Init(serviceName)
	// defer func() {
	// 	if err := tp.Shutdown(context.Background()); err != nil {
	// 		log.Printf("Error shutting down tracer provider: %v", err)
	// 	}
	// }()
	tracer = otel.Tracer(serviceName)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	// initDB()
	go setupServer()

	<-sigint
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server shutdown failed")
	}
}
