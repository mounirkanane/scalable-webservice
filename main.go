package main

import  (
	"net/http"
	"log"
	"encoding/json"
	"time"
	"strings"
)

type book struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
}

var books = []book{
	{ID: "1", Title: "The Go Programming Language", Author: "Alan Donovan", Price: 39.99},
	{ID: "2", Title: "Clean Code", Author: "Robert Martin", Price: 32.50},
	{ID: "3", Title: "System Design Interview", Author: "Alex Xu", Price: 28.95},
	{ID: "4", Title: "Dopamine Nation", Author: "Anne Lembke", Price: 20.00},
}

type BookDetailsResponse struct {
    BookID   string                 `json:"book_id"`
    Metadata map[string]interface{} `json:"metadata"`
    Pricing  map[string]interface{} `json:"pricing"`
    Inventory map[string]interface{} `json:"inventory"`
    Reviews  map[string]interface{} `json:"reviews"`
    Duration int64                  `json:"duration"`
}

func main() {
	http.HandleFunc("/api/books", booksHandler)
	http.HandleFunc("/api/books/", bookDetailHandler)
	log.Println("Starting server on http://localhost:8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("FATAL: error while starting server:", err)
    }
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
	// Validate the HTTP method
	if r.Method != http.MethodGet {
		log.Printf("Method %s not allowed for %s", r.Method, r.URL.Path)
		// Sets status code and sends an error message
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // Client errors are 400s, in this case 405
		return                                                           // Need this to prevent rest of handler from executing after error
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")

	// Encode and stream books as a JSON response, the JSON is sent across the network as it's generated
	err := json.NewEncoder(w).Encode(books)
	if err != nil {
		log.Printf("Error occured while encoding JSON: %v", err)
		return
	}

	// Log successful operation
	log.Printf("Successfully returned %d books to %s", len(books), r.RemoteAddr)
}

func bookDetailHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/") // {"", "api", "books", "123", "details"}
	// Verify URL format
	if len(pathParts) < 5 || pathParts[4] != "details" {
		http.Error(w, "Invalid URL Format. Expected /api/books/{id}/details", http.StatusBadRequest)
		return
	}

	bookID := pathParts[3]
	log.Printf("Processing book details request for ID: %s", bookID)

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "sequential"
	}

	log.Printf("Processing book details request for ID: %s using %s mode", bookID, mode)

	switch mode {
		case "sequential":
			handleSequentialBookDetails(w, r, bookID)
		case "concurrent": 
			handleConcurrentBookDetails(w, r, bookID)
		default:
			http.Error(w, "Invalid mode. Use 'sequential' or 'concurrent'", http.StatusBadRequest)
	}
}

func handleConcurrentBookDetails(w http.ResponseWriter, r *http.Request, bookID string) {
	startTime := time.Now()

	// Create channels to receive results from each service call
	metadataChannel := make(chan map[string]interface{})
	pricingChannel := make(chan map[string]interface{})
	inventoryChannel := make(chan map[string]interface{})
	reviewsChannel := make(chan map[string]interface{})

	// Launch goroutine for each channel
	go func() {
		result := fetchBookMetadata(bookID)
		metadataChannel <- result
	}() // calls the anon function

	go func() {
		result := fetchBookPricing(bookID)
		pricingChannel <- result
	}()

	go func() {
		result := fetchBookInventory(bookID)
		inventoryChannel <- result
	}()

	go func() {
		result := fetchBookReviews(bookID)
		reviewsChannel <- result
	}()

	response := BookDetailsResponse {
		BookID: bookID,
		Metadata: <-metadataChannel,
		Pricing: <-pricingChannel,
		Inventory: <-inventoryChannel,
		Reviews: <-reviewsChannel,
		Duration: time.Since(startTime).Milliseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(response)

	log.Printf("Concurrent processing completed in %v", time.Since(startTime))
}



func handleSequentialBookDetails(w http.ResponseWriter, r *http.Request, bookID string) {
	startTime := time.Now()

	// Simulate fetching from metadata service
    metadata := fetchBookMetadata(bookID)
    
    // Simulate fetching from pricing service
    pricing := fetchBookPricing(bookID)
    
    // Simulate fetching from inventory service
    inventory := fetchBookInventory(bookID)
    
    // Simulate fetching from reviews service
    reviews := fetchBookReviews(bookID)

	response := BookDetailsResponse {
		BookID:   bookID,
		Metadata: metadata,
		Pricing:  pricing,
		Inventory: inventory,
		Reviews:  reviews,
		Duration: time.Since(startTime).Milliseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(response)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // Use two spaces for indentation
	encoder.Encode(response)

	log.Printf("Sequential processing completed in %v", time.Since(startTime))
} 

// Simulate book metadata fetching
func fetchBookMetadata(bookID string) map[string]interface{} {

	time.Sleep(80 * time.Millisecond)

	return map[string]interface{}{
        "title":         "Sample Book Title",
        "author":        "Sample Author",
        "isbn":          "978-0123456789",
        "publish_date":  "2023-01-15",
        "description":   "A detailed description of the book",
    }
}

// Simulate fetching price info
func fetchBookPricing(bookID string) map[string]interface{} {
	time.Sleep(120 * time.Millisecond)

	return map[string]interface{}{
        "price":         29.99,
        "currency":      "USD",
        "discount":      0.10,
        "sale_price":    26.99,
        "promotion":     "Limited time offer",
    }
}

// Simulate fetching inventory status
func fetchBookInventory(bookID string) map[string]interface{} {
	time.Sleep(150 * time.Millisecond)

	return map[string]interface{}{
        "in_stock":      true,
        "quantity":      42,
        "warehouse":     "East Coast Distribution",
        "shipping_time": "2-3 business days",
    }
}

// Simulate fetching customer reviews
func fetchBookReviews(bookID string) map[string]interface{} {
	time.Sleep(100 * time.Millisecond)

	return map[string]interface{}{
        "average_rating": 4.3,
        "total_reviews":  127,
        "recent_review":  "Great book, highly recommended!",
        "rating_breakdown": map[string]int{
            "5_star": 65,
            "4_star": 32,
            "3_star": 20,
            "2_star": 7,
            "1_star": 3,
        },
	}
}