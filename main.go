package main

import (
	"log"
	"net/http"
)

func main() {
	// Initialize database connection and schema
	err := InitializeDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Ensure database connection closes when application exits
	defer func() {
		if err := CloseDatabase(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Register HTTP route handlers
	http.HandleFunc("/api/books", BooksHandler)       // Simple books list
	http.HandleFunc("/api/books/", BookDetailHandler) // Detailed book information

	// Start HTTP server
	log.Println("Starting server on http://localhost:8080")
	log.Println("Available endpoints:")
	log.Println("  GET /api/books - List all books")
	log.Println("  GET /api/books/{id}/details?mode=sequential - Sequential operations")
	log.Println("  GET /api/books/{id}/details?mode=concurrent - Concurrent operations")
	log.Println("  Optional: &user_id=demo_user for personalized recommendations")
	log.Println("")
	log.Println("Operations include:")
	log.Println("  • Database queries for metadata, pricing, inventory, reviews")
	log.Println("  • External API call to api.quotable.io for recommendations")
	log.Println("")
	log.Println("This demonstrates the difference between sequential and concurrent coordination")
	log.Println("when mixing fast database operations with slower external API calls.")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("FATAL: error while starting server:", err)
	}
}
