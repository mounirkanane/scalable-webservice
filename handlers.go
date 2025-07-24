package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// BooksHandler handles requests to /api/books (returns simple list of books)
func BooksHandler(w http.ResponseWriter, r *http.Request) {
	// Validate the HTTP method
	if r.Method != http.MethodGet {
		log.Printf("Method %s not allowed for %s", r.Method, r.URL.Path)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode and stream books as a JSON response
	err := json.NewEncoder(w).Encode(books)
	if err != nil {
		log.Printf("Error occurred while encoding JSON: %v", err)
		return
	}

	// Log successful operation
	log.Printf("Successfully returned %d books to %s", len(books), r.RemoteAddr)
}

// BookDetailHandler handles requests to /api/books/{id}/details with mode selection
func BookDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Parse URL path to extract book ID
	pathParts := strings.Split(r.URL.Path, "/") // {"", "api", "books", "123", "details"}

	// Verify URL format
	if len(pathParts) < 5 || pathParts[4] != "details" {
		http.Error(w, "Invalid URL Format. Expected /api/books/{id}/details", http.StatusBadRequest)
		return
	}

	// Extract book ID from URL
	bookID := pathParts[3]
	log.Printf("Processing book details request for ID: %s", bookID)

	// Check query parameter for processing mode (default to sequential)
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "sequential"
	}

	log.Printf("Processing book details request for ID: %s using %s mode", bookID, mode)

	// Route to appropriate handler based on mode
	switch mode {
	case "sequential":
		handleSequentialBookDetails(w, r, bookID)
	case "concurrent":
		handleConcurrentBookDetails(w, r, bookID)
	default:
		http.Error(w, "Invalid mode. Use 'sequential' or 'concurrent'", http.StatusBadRequest)
	}
}

// handleSequentialBookDetails processes database queries one after another
func handleSequentialBookDetails(w http.ResponseWriter, r *http.Request, bookID string) {
	startTime := time.Now()

	// Sequential approach: call each database query one at a time
	metadata := FetchBookMetadata(bookID)
	pricing := FetchBookPricing(bookID)
	inventory := FetchBookInventory(bookID)
	reviews := FetchBookReviews(bookID)

	// Build comprehensive response
	response := BookDetailsResponse{
		BookID:    bookID,
		Metadata:  metadata,
		Pricing:   pricing,
		Inventory: inventory,
		Reviews:   reviews,
		Duration:  time.Since(startTime).Milliseconds(),
	}

	// Send JSON response with pretty printing
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(response)

	log.Printf("Sequential processing completed in %v", time.Since(startTime))
}

// handleConcurrentBookDetails processes database queries concurrently using goroutines
func handleConcurrentBookDetails(w http.ResponseWriter, r *http.Request, bookID string) {
	startTime := time.Now()

	// Create channels to receive results from each concurrent database query
	metadataChannel := make(chan map[string]interface{})
	pricingChannel := make(chan map[string]interface{})
	inventoryChannel := make(chan map[string]interface{})
	reviewsChannel := make(chan map[string]interface{})

	// Launch concurrent goroutines for each database query
	go func() {
		result := FetchBookMetadata(bookID)
		metadataChannel <- result
	}()

	go func() {
		result := FetchBookPricing(bookID)
		pricingChannel <- result
	}()

	go func() {
		result := FetchBookInventory(bookID)
		inventoryChannel <- result
	}()

	go func() {
		result := FetchBookReviews(bookID)
		reviewsChannel <- result
	}()

	// Collect results from all channels (fan-in coordination)
	// This blocks until all goroutines complete and send their results
	response := BookDetailsResponse{
		BookID:    bookID,
		Metadata:  <-metadataChannel,
		Pricing:   <-pricingChannel,
		Inventory: <-inventoryChannel,
		Reviews:   <-reviewsChannel,
		Duration:  time.Since(startTime).Milliseconds(),
	}

	// Send JSON response with pretty printing
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(response)

	log.Printf("Concurrent processing completed in %v", time.Since(startTime))
}
