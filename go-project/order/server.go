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

const serviceName = "order-service"

var (
	// db       datastore.DB
	srv      *http.Server
	orderUrl string
	userUrl  string
	tracer   trace.Tracer
)

func setupServer() {
	router := mux.NewRouter()
	router.HandleFunc("/orders/", getOrders).Methods(http.MethodGet, http.MethodOptions)
	// router.Use(utils.LoggingMW)
	router.Use(otelmux.Middleware(serviceName))
	router.Use(utils.MetricMiddleware)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost},
	})

	srv = &http.Server{
		Addr:    ":8081",
		Handler: c.Handler(router),
	}

	log.Printf("Order service running at: :8081")
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to setup http server: %v", err)
	}
}

func main() {
	utils.InitTraceProvider()
	utils.InitMetricProvider(utils.InitServerMeter)
	orderUrl = os.Getenv("ORDER_URL")
	userUrl = os.Getenv("USER_URL")

	tracer = otel.Tracer(serviceName)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	go setupServer()
	<-sigint
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server shutdown failed")
	}
}
