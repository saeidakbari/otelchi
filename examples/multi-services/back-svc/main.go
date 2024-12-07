package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/riandyrn/otelchi"
	"github.com/riandyrn/otelchi/examples/multi-services/utils"
	"go.opentelemetry.io/otel/trace"
)

const (
	addr        = ":8091"
	serviceName = "back-svc"
)

func main() {
	// init tracer provider
	tracer, err := utils.NewTracer(serviceName)
	if err != nil {
		log.Fatalf("unable to initialize tracer provider due: %v", err)
	}

	// initialize metrics middleware
	metricsMiddleware, err := utils.NewMetricProvider(serviceName)
	if err != nil {
		log.Fatalf("unable to initialize metrics due: %v", err)
	}

	// define router
	r := chi.NewRouter()
	r.Use(otelchi.Middleware(serviceName, otelchi.WithChiRoutes(r)))
	r.Use(metricsMiddleware)
	r.Get("/", utils.HealthCheckHandler)
	r.Get("/name", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(generateName(r.Context(), tracer)))
	})
	r.Handle("/metrics", promhttp.Handler())
	log.Printf("back service is listening on %v", addr)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("unable to execute server due: %v", err)
	}
}

func generateName(ctx context.Context, tracer trace.Tracer) string {
	_, span := tracer.Start(ctx, "generateName")
	defer span.End()

	rndNum := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100000)
	return fmt.Sprintf("user_%v", rndNum)
}
