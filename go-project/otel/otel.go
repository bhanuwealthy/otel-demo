package otel

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	//make sure some constructors are executed only once
	once sync.Once
	//meter provider is common across the process
	MeterProvider   *sdk.MeterProvider
	mpcounter       int
	metricResources *resource.Resource
	metricExporter  *otlpmetricgrpc.Exporter
	nodeIP          attribute.Key = "node.ip"
	GithubHash      string
	TracerProvider  *sdktrace.TracerProvider
)

type MeterConstructor func(meterProvider *sdk.MeterProvider)

func InitMetricProvider(meterConstructors ...MeterConstructor) func(context.Context) error {
	once.Do(initMeterProvider)
	// Instantiate the OTLP GRPC exporter
	for _, m := range meterConstructors {
		m(MeterProvider)
	}
	fmt.Println("mp counter ", mpcounter)
	return MeterProvider.Shutdown
}

func init() {
	if os.Getenv("OTEL_TRACE_ENABLED") == "1" {
		InitTraceProvider()
	}
}

func initMeterProvider() {
	mpcounter++
	var err error
	ctx := context.Background()
	otelEndpoint := "http://otel-agent-service.monitoring:4317"
	if agentAddrFromEnv := os.Getenv("OTEL_AGENT_ADDRESS"); agentAddrFromEnv != "" {
		otelEndpoint = agentAddrFromEnv
	}
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", otelEndpoint)
	//default resource definition
	metricResources = newResource()
	// Instantiate the OTLP GRPC exporter
	metricExporter, err = otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	if err != nil {
		log.Fatalf("new otlp metric grpc exporter failed: %v", err)
	}
	MeterProvider = sdk.NewMeterProvider(
		sdk.WithResource(metricResources),
		sdk.WithReader(sdk.NewPeriodicReader(metricExporter, sdk.WithInterval(time.Second*10))),
	)
	// return meterProvider.Shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		fmt.Println("shut down meter provider")
		MeterProvider.Shutdown(ctx)
	}()
}

func newResource() *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(os.Getenv("APP_NAME")),
		semconv.ServiceVersionKey.String(GithubHash),
		semconv.ContainerRuntimeKey.String(runtime.Version()),
		semconv.ContainerNameKey.String(os.Getenv("APP_NAME")),
		semconv.ContainerIDKey.String(os.Getenv("POD_NAME")),
		semconv.ServiceNamespaceKey.String(os.Getenv("POD_NAMESPACE")),
		semconv.DeploymentEnvironmentKey.String(os.Getenv("APP_ENV")),
		nodeIP.String(os.Getenv("NODE_IP")),
	)
}

func InitTraceProvider() {
	otelEndpoint := "http://otel-agent-service.monitoring:4317"
	if agentAddrFromEnv := os.Getenv("OTEL_AGENT_ADDRESS"); agentAddrFromEnv != "" {
		otelEndpoint = agentAddrFromEnv
	}
	// otelEndpoint := "http://localhost:4317"
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", otelEndpoint)
	traceExporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
		),
	)
	if err != nil {
		log.Fatalf("error initializing otel tracer %v", err)
	}

	TracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(traceExporter)),
		// sdktrace.WithSyncer(traceExporter),
		sdktrace.WithResource(newResource()),
	)
	otel.SetTracerProvider(TracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		if err := TracerProvider.Shutdown(context.Background()); err != nil {
			log.Printf("error shutting down tracer provider: %v", err)
		}
		os.Exit(0)
	}()
}
