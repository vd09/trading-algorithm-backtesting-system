package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Create a new counter metric
	opsProcessed := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})

	// Register the metric
	prometheus.MustRegister(opsProcessed)

	go func() {
		// Expose the registered metrics via HTTP
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()

	ctx := context.Background()
	// Increment the counter
	opsProcessed.Set(1)
	opsProcessed.SetToCurrentTime()
	fmt.Println("vdji1")
	time.Sleep(30 * time.Second)
	fmt.Println("vdji2")
	opsProcessed.Set(2)
	fmt.Println("vdji3")

	<-ctx.Done()
}
