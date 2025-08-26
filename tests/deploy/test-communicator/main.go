package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "test-communicator/proto"
)

type Config struct {
	Port        int    `json:"port"`
	ServiceName string `json:"service_name"`
	TargetURL   string `json:"target_url"`
	Protocol    string `json:"protocol"`    // "http", "grpc", or "tcp"
	TargetHost  string `json:"target_host"` // For non-HTTP protocols
	TargetPort  int    `json:"target_port"` // For non-HTTP protocols
}

type App struct {
	config     Config
	router     *mux.Router
	httpServer *http.Server
	grpcServer *grpc.Server
	tcpServer  net.Listener
	requests   prometheus.Counter
	stopCh     chan struct{}
}

// gRPC server implementation
type testCommunicatorServer struct {
	pb.UnimplementedTestCommunicatorServer
	serviceName string
}

func (s *testCommunicatorServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Status:    "healthy",
		Service:   s.serviceName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *testCommunicatorServer) GetData(ctx context.Context, req *pb.DataRequest) (*pb.DataResponse, error) {
	return &pb.DataResponse{
		Message:   "Data retrieved successfully via gRPC",
		Service:   s.serviceName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *testCommunicatorServer) CallTarget(ctx context.Context, req *pb.TargetRequest) (*pb.TargetResponse, error) {
	return &pb.TargetResponse{
		Message:   "Target called successfully via gRPC",
		Service:   s.serviceName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

type DataResponse struct {
	Message   string `json:"message"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

type UserResponse struct {
	ID      string `json:"id"`
	Service string `json:"service"`
	Created string `json:"created"`
}

func NewApp() *App {
	config := Config{
		Port:        getEnvAsInt("PORT", 8080),
		ServiceName: getEnv("SERVICE_NAME", "test-communicator"),
		TargetURL:   getEnv("TARGET_URL", ""),
		Protocol:    getEnv("PROTOCOL", "http"),
		TargetHost:  getEnv("TARGET_HOST", ""),
		TargetPort:  getEnvAsInt("TARGET_PORT", 8080),
	}

	app := &App{
		config: config,
		router: mux.NewRouter(),
		stopCh: make(chan struct{}),
	}

	// Initialize Prometheus metrics
	app.requests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "requests_total",
		Help: "Total number of requests",
	})
	prometheus.MustRegister(app.requests)

	// Setup HTTP routes if HTTP protocol is enabled
	if config.Protocol == "http" || config.Protocol == "all" {
		app.setupHTTPRoutes()
	}

	return app
}

func (a *App) setupHTTPRoutes() {
	// Health endpoint
	a.router.HandleFunc("/health", a.healthHandler).Methods("GET")

	// API endpoints
	a.router.HandleFunc("/api/data", a.dataHandler).Methods("GET")
	a.router.HandleFunc("/api/users/{id}", a.userHandler).Methods("GET")
	a.router.HandleFunc("/api/call-target", a.callTargetHandler).Methods("GET")

	// Metrics endpoint
	a.router.Handle("/metrics", promhttp.Handler())

	// Root endpoint
	a.router.HandleFunc("/", a.rootHandler).Methods("GET")

	// Add middleware for request counting
	a.router.Use(a.requestCounterMiddleware)
}

func (a *App) requestCounterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.requests.Inc()
		next.ServeHTTP(w, r)
	})
}

func (a *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Service:   a.config.ServiceName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *App) dataHandler(w http.ResponseWriter, r *http.Request) {
	response := DataResponse{
		Message:   "Data retrieved successfully",
		Service:   a.config.ServiceName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *App) userHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	response := UserResponse{
		ID:      userID,
		Service: a.config.ServiceName,
		Created: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *App) callTargetHandler(w http.ResponseWriter, r *http.Request) {
	if a.config.TargetURL == "" {
		http.Error(w, "No target URL configured", http.StatusBadRequest)
		return
	}

	log.Printf("Making HTTP request to target: %s", a.config.TargetURL)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(a.config.TargetURL + "/health")
	if err != nil {
		log.Printf("Error calling target: %v", err)
		http.Error(w, fmt.Sprintf("Error calling target: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		http.Error(w, fmt.Sprintf("Error reading response: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":         "Successfully called target service",
		"service":         a.config.ServiceName,
		"target_url":      a.config.TargetURL,
		"target_status":   resp.StatusCode,
		"target_response": string(body),
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *App) rootHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service":   a.config.ServiceName,
		"endpoints": []string{"/health", "/api/data", "/api/users/{id}", "/api/call-target", "/metrics"},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *App) Start() error {
	log.Printf("Starting %s server with protocol: %s on port %d", a.config.ServiceName, a.config.Protocol, a.config.Port)
	log.Printf("Target URL: %s, Target Host: %s, Target Port: %d", a.config.TargetURL, a.config.TargetHost, a.config.TargetPort)

	switch a.config.Protocol {
	case "http":
		return a.startHTTPServer()
	case "grpc":
		return a.startGRPCServer()
	case "tcp":
		return a.startTCPServer()
	case "all":
		// Start all servers
		go a.startGRPCServer()
		go a.startTCPServer()
		return a.startHTTPServer()
	default:
		return fmt.Errorf("unsupported protocol: %s", a.config.Protocol)
	}
}

func (a *App) startHTTPServer() error {
	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.Port),
		Handler: a.router,
	}

	// Start periodic client requests if target is configured
	if a.config.TargetURL != "" || a.config.TargetHost != "" {
		go a.startPeriodicRequests()
	}

	// Start server in a goroutine
	go func() {
		log.Printf("HTTP server listening on :%d", a.config.Port)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	return nil
}

func (a *App) startGRPCServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	a.grpcServer = grpc.NewServer()
	pb.RegisterTestCommunicatorServer(a.grpcServer, &testCommunicatorServer{
		serviceName: a.config.ServiceName,
	})

	// Start periodic client requests if target is configured
	if a.config.TargetHost != "" {
		go a.startPeriodicRequests()
	}

	go func() {
		log.Printf("gRPC server listening on :%d", a.config.Port)
		if err := a.grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed to start: %v", err)
		}
	}()

	return nil
}

func (a *App) startTCPServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	a.tcpServer = lis

	// Start periodic client requests if target is configured
	if a.config.TargetHost != "" {
		go a.startPeriodicRequests()
	}

	go func() {
		log.Printf("TCP server listening on :%d", a.config.Port)
		for {
			conn, err := lis.Accept()
			if err != nil {
				log.Printf("TCP server accept error: %v", err)
				return
			}
			go a.handleTCPConnection(conn)
		}
	}()

	return nil
}

func (a *App) handleTCPConnection(conn net.Conn) {
	defer conn.Close()
	a.requests.Inc()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("TCP received: %s", line)

		var response string
		switch {
		case strings.Contains(line, "health"):
			response = fmt.Sprintf(`{"status":"healthy","service":"%s","timestamp":"%s"}`,
				a.config.ServiceName, time.Now().UTC().Format(time.RFC3339))
		case strings.Contains(line, "data"):
			response = fmt.Sprintf(`{"message":"Data retrieved successfully via TCP","service":"%s","timestamp":"%s","items":["item1","item2","item3"],"count":3,"active":true}`,
				a.config.ServiceName, time.Now().UTC().Format(time.RFC3339))
		default:
			response = fmt.Sprintf(`{"message":"TCP server response","service":"%s","timestamp":"%s"}`,
				a.config.ServiceName, time.Now().UTC().Format(time.RFC3339))
		}

		conn.Write([]byte(response + "\n"))
	}
}

func (a *App) startPeriodicRequests() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Printf("Starting periodic requests to target every minute")

	// Make an initial request after 30 seconds to avoid startup race conditions
	time.Sleep(30 * time.Second)
	a.makeTargetRequest()

	for {
		select {
		case <-ticker.C:
			a.makeTargetRequest()
		case <-a.stopCh:
			log.Println("Stopping periodic requests")
			return
		}
	}
}

func (a *App) makeTargetRequest() {
	switch a.config.Protocol {
	case "http":
		a.makeHTTPTargetRequest()
	case "grpc":
		a.makeGRPCTargetRequest()
	case "tcp":
		a.makeTCPTargetRequest()
	case "all":
		// Make requests with all protocols
		a.makeHTTPTargetRequest()
		a.makeGRPCTargetRequest()
		a.makeTCPTargetRequest()
	}
}

func (a *App) makeHTTPTargetRequest() {
	if a.config.TargetURL == "" {
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	log.Printf("Making periodic HTTP request to target: %s", a.config.TargetURL)

	resp, err := client.Get(a.config.TargetURL + "/health")
	if err != nil {
		log.Printf("Error in periodic HTTP request to target: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading periodic HTTP response: %v", err)
		return
	}

	log.Printf("Periodic HTTP request successful - Status: %d, Response: %s", resp.StatusCode, string(body))
}

func (a *App) makeGRPCTargetRequest() {
	if a.config.TargetHost == "" {
		return
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", a.config.TargetHost, a.config.TargetPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Error connecting to gRPC target: %v", err)
		return
	}
	defer conn.Close()

	client := pb.NewTestCommunicatorClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Making periodic gRPC request to target: %s:%d", a.config.TargetHost, a.config.TargetPort)

	resp, err := client.Health(ctx, &pb.HealthRequest{})
	if err != nil {
		log.Printf("Error in periodic gRPC request: %v", err)
		return
	}

	log.Printf("Periodic gRPC request successful - Response: %v", resp)
}

func (a *App) makeTCPTargetRequest() {
	if a.config.TargetHost == "" {
		return
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", a.config.TargetHost, a.config.TargetPort), 10*time.Second)
	if err != nil {
		log.Printf("Error connecting to TCP target: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Making periodic TCP request to target: %s:%d", a.config.TargetHost, a.config.TargetPort)

	// Send health check request
	_, err = conn.Write([]byte("health\n"))
	if err != nil {
		log.Printf("Error writing to TCP connection: %v", err)
		return
	}

	// Read response
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		response := scanner.Text()
		log.Printf("Periodic TCP request successful - Response: %s", response)
	} else {
		log.Printf("Error reading TCP response: %v", scanner.Err())
	}
}

func (a *App) Stop() error {
	log.Println("Shutting down servers...")

	// Signal periodic requests to stop
	close(a.stopCh)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var errors []error

	// Stop HTTP server
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("HTTP server shutdown error: %v", err))
		}
	}

	// Stop gRPC server
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}

	// Stop TCP server
	if a.tcpServer != nil {
		if err := a.tcpServer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("TCP server shutdown error: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("server shutdown errors: %v", errors)
	}

	return nil
}

func main() {
	app := NewApp()

	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := app.Stop(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
