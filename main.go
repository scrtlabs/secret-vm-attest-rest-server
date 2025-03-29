package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"secret-vm-attest-rest-server/pkg"
	"syscall"
	"time"
)

func main() {
	// Define command-line flags with default values coming from pkg.
	secure := flag.Bool("secure", pkg.Secure, "Enable HTTPS")
	port := flag.Int("port", pkg.Port, "Port to listen on")
	ip := flag.String("ip", pkg.RESTServerIP, "IP address to bind to")
	flag.Parse()

	// Construct the address string.
	addr := fmt.Sprintf("%s:%d", *ip, *port)

	// Create a new HTTP request multiplexer.
	mux := http.NewServeMux()

	// Register endpoint handlers.
	mux.HandleFunc("/status", pkg.StatusHandler)
	mux.HandleFunc("/gpu", pkg.MakeAttestationFileHandler(pkg.GPUAttestationFile, "GPU"))
	mux.HandleFunc("/cpu", pkg.MakeAttestationFileHandler(pkg.CPUAttestationFile, "CPU"))
	mux.HandleFunc("/self", pkg.MakeAttestationFileHandler(pkg.SelfAttestationFile, "Self"))

	// Apply middleware chain - order matters here
	// First CORS, then security headers, and finally logging
	handler := pkg.LoggingMiddleware(
		pkg.SecurityHeadersMiddleware(
			pkg.CORSMiddleware(mux),
		),
	)

	// Create a new server with reasonable timeout settings
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a goroutine so it doesn't block
	go func() {
		log.Printf("Server starting on %s, secure: %v", addr, *secure)
		var err error

		if *secure {
			// Use configuration values for certificate paths
			certPath := pkg.CertPath
			keyPath := pkg.KeyPath

			// Check if certificate and key files exist.
			if _, err := os.Stat(certPath); os.IsNotExist(err) {
				log.Fatalf("SSL certificate file not found at %s", certPath)
			}
			if _, err := os.Stat(keyPath); os.IsNotExist(err) {
				log.Fatalf("SSL key file not found at %s", keyPath)
			}

			// Start the HTTPS server.
			err = server.ListenAndServeTLS(certPath, keyPath)
		} else {
			// Start the HTTP server.
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Set up channel for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal
	<-stop

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	log.Println("Server is shutting down...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}
