package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"secretai-attest-rest/pkg"
)

func main() {
	// Define command-line flags for secure mode, port, and IP address.
	secure := flag.Bool("secure", true, "Enable HTTPS")
	port := flag.Int("port", 29343, "Port to listen on")
	ip := flag.String("ip", pkg.GetEnv("SECRETAI_REST_SERVER_IP", "0.0.0.0"), "IP address to bind to")
	flag.Parse()

	// Construct the address string.
	addr := fmt.Sprintf("%s:%d", *ip, *port)

	// Create a new HTTP request multiplexer.
	mux := http.NewServeMux()

	// Register endpoint handlers.
	mux.HandleFunc("/status", pkg.StatusHandler)
	mux.HandleFunc("/attestation", pkg.AttestationHandler)
	mux.HandleFunc("/gpu", pkg.MakeAttestationFileHandler(pkg.GPUAttestationFile, "GPU"))
	mux.HandleFunc("/cpu", pkg.MakeAttestationFileHandler(pkg.CPUAttestationFile, "CPU"))
	mux.HandleFunc("/self", pkg.MakeAttestationFileHandler(pkg.SelfAttestationFile, "Self"))

	// Wrap the mux with logging middleware.
	handler := pkg.LoggingMiddleware(mux)

	log.Printf("Server starting on %s, secure: %v", addr, *secure)
	if *secure {
		// Paths to the SSL certificate and key.
		certPath := "cert/ssl_cert.pem"
		keyPath := "cert/ssl_key.pem"
		// Check if certificate and key files exist.
		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			log.Fatalf("SSL certificate file not found at %s", certPath)
		}
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			log.Fatalf("SSL key file not found at %s", keyPath)
		}
		// Start the HTTPS server.
		err := http.ListenAndServeTLS(addr, certPath, keyPath, handler)
		if err != nil {
			log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	} else {
		// Start the HTTP server.
		err := http.ListenAndServe(addr, handler)
		if err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}
}
