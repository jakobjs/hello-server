package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	serverAddress         = ":8080"
	healthFailureEvery    = 30 * time.Minute
	healthFailureDuration = 10 * time.Minute
	readyAfter            = 2 * time.Minute
)

type healthChecker struct {
	startedAt time.Time
	now       func() time.Time
}

type readinessChecker struct {
	startedAt time.Time
	now       func() time.Time
}

func newHealthChecker(startedAt time.Time) healthChecker {
	return healthChecker{
		startedAt: startedAt,
		now:       time.Now,
	}
}

func newReadinessChecker(startedAt time.Time) readinessChecker {
	return readinessChecker{
		startedAt: startedAt,
		now:       time.Now,
	}
}

func (checker healthChecker) isHealthy() bool {
	elapsed := checker.now().Sub(checker.startedAt)
	if elapsed < healthFailureEvery {
		return true
	}

	return elapsed%healthFailureEvery >= healthFailureDuration
}

func (checker readinessChecker) isReady() bool {
	return checker.now().Sub(checker.startedAt) >= readyAfter
}

func RootServer(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received request for %s\n", r.URL.Path)
	fmt.Fprintf(w, "Welcome to the root path! Use /slow for a delayed response. Use /hello/{name} to get a personalized greeting.")
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received request for %s\n", r.URL.Path)
	name := strings.TrimPrefix(r.URL.Path, "/hello/")
	name = strings.Trim(name, "/")
	if name == "" {
		http.Error(w, "please provide a name in /hello/{name}", http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Hello, %s!", name)
}

func SlowHelloServer(w http.ResponseWriter, r *http.Request) {
	delaySeconds := 5
	if raw := r.URL.Query().Get("seconds"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 && parsed <= 10 {
			delaySeconds = parsed
		}
	}

	delay := time.Duration(delaySeconds) * time.Second
	start := time.Now()
	fmt.Printf("Started /slow request with %ds delay\n", delaySeconds)
	time.Sleep(delay)

	fmt.Fprintf(w, "Slow response after %v\n", time.Since(start).Round(100*time.Millisecond))
}

func HealthServer(checker healthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Received request for %s\n", r.URL.Path)
		if checker.isHealthy() {
			fmt.Fprintln(w, "ok")
			return
		}

		http.Error(w, "unhealthy", http.StatusServiceUnavailable)
	}
}

func ReadyServer(checker readinessChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Received request for %s\n", r.URL.Path)
		if checker.isReady() {
			fmt.Fprintln(w, "ready")
			return
		}

		http.Error(w, "not ready", http.StatusServiceUnavailable)
	}
}

func main() {
	fmt.Println("Starting web server")
	startTime := time.Now()
	health := newHealthChecker(startTime)
	ready := newReadinessChecker(startTime)

	http.HandleFunc("/", RootServer)
	http.HandleFunc("/hello/", HelloServer)
	http.HandleFunc("/slow", SlowHelloServer)
	http.HandleFunc("/healthz", HealthServer(health))
	http.HandleFunc("/ready", ReadyServer(ready))

	server := &http.Server{Addr: serverAddress, Handler: nil}

	go func() {
		fmt.Printf("Web server is running on http://localhost%s ...\n", serverAddress)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("Server stopped:", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	fmt.Println("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Println("Graceful shutdown failed:", err)
	}
}
