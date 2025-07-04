package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"secret-vm-attest-rest-server/pkg"
	"syscall"
	"time"
)

// 1. Embed everything under pkg/html/images/
//
//go:embed pkg/html/images/*
var embeddedImages embed.FS

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

	// carve out a sub-FS whose root is pkg/html/images
	imagesSub, err := fs.Sub(embeddedImages, "pkg/html/images")
	if err != nil {
		log.Fatalf("failed to create sub FS for images: %v", err)
	}

	// now imagesSub has "favicon.png", etc. at its root
	imageDir := http.FileServer(http.FS(imagesSub))
	mux.Handle("/images/", http.StripPrefix("/images/", imageDir))

	mux.HandleFunc("/status", pkg.StatusHandler)
	// Register endpoints returning attestation text.
	mux.HandleFunc("/gpu", pkg.MakeAttestationFileHandler(pkg.GPUAttestationFile, "GPU"))
	mux.HandleFunc("/cpu", pkg.MakeAttestationFileHandler(pkg.CPUAttestationFile, "CPU"))
	mux.HandleFunc("/self", pkg.MakeAttestationFileHandler(pkg.SelfAttestationFile, "Self"))

	// Register endpoints returning attestation as rendered HTML.
	mux.HandleFunc("/gpu.html", pkg.MakeAttestationHTMLHandler(pkg.GPUAttestationFile, "GPU"))
	mux.HandleFunc("/cpu.html", pkg.MakeAttestationHTMLHandler(pkg.CPUAttestationFile, "CPU"))
	mux.HandleFunc("/self.html", pkg.MakeAttestationHTMLHandler(pkg.SelfAttestationFile, "Self"))

	// VM logs endpoints
	mux.HandleFunc("/logs", pkg.MakeVMLogsHandler(*secure))
	mux.HandleFunc("/logs.html", pkg.MakeVMLiveLogsHandler())

	// New endpoints for docker-compose
	mux.HandleFunc("/docker-compose", pkg.MakeDockerComposeFileHandler())
	mux.HandleFunc("/docker-compose.html", pkg.MakeDockerComposeHTMLHandler())

	// New endpoints for resources
	mux.HandleFunc("/resources", pkg.MakeResourcesHandler())
	mux.HandleFunc("/resources.html", pkg.MakeResourcesHTMLHandler())

	// New endpoints for image updates
	mux.HandleFunc("/vm_updates", pkg.MakeVMUpdatesHandler())
	mux.HandleFunc("/vm_updates.html", pkg.MakeVMUpdatesHTMLHandler())

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
